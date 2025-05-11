package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/ivanov-nikolay/distributed_calculator/internal/app/entity"
	"github.com/ivanov-nikolay/distributed_calculator/internal/app/usecase"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/calculation"
	"github.com/ivanov-nikolay/distributed_calculator/pkg/jwt"
)

type Handler struct {
	exprUC  *usecase.ExpressionUseCase
	userUC  *usecase.UserUseCase
	taskUC  *usecase.TaskUseCase
	jwtUtil *jwt.JWTUtil
}

func NewHandler(
	exprUC *usecase.ExpressionUseCase,
	userUC *usecase.UserUseCase,
	taskUC *usecase.TaskUseCase,
	jwtUtil *jwt.JWTUtil) *Handler {
	return &Handler{
		exprUC:  exprUC,
		userUC:  userUC,
		taskUC:  taskUC,
		jwtUtil: jwtUtil,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/register", h.RegisterHandler)
	mux.HandleFunc("POST /api/v1/login", h.LoginHandler)

	mux.Handle("POST /api/v1/calculate", h.AuthMiddleware(http.HandlerFunc(h.CalculateHandler)))
	mux.Handle("GET /api/v1/expressions", h.AuthMiddleware(http.HandlerFunc(h.ExpressionsHandler)))
	mux.Handle("GET /api/v1/expressions/{id}", h.AuthMiddleware(http.HandlerFunc(h.ExpressionByIDHandler)))

	mux.HandleFunc("GET /internal/task", h.GetTaskHandler)
	mux.HandleFunc("POST /internal/task", h.PostTaskHandler)

	mux.HandleFunc("/", h.NotFoundHandler)
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.userUC.Create(r.Context(), req.Login, req.Password); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "user created"})
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	userID, err := h.userUC.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.jwtUtil.GenerateToken(userID, req.Login)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *Handler) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	log.Println("CalculateHandler from context userID, ok := r.Context().Value(\"userID\").(int):", userID, ok)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	expr, err := h.exprUC.Add(r.Context(), userID, req.Expression)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	ast, err := calculation.ParseAST(req.Expression)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	exprID := strconv.Itoa(expr)

	entityAST := convertASTNode(ast)

	expression := &entity.Expression{
		ID:     exprID,
		Expr:   req.Expression,
		Status: "pending",
		AST:    entityAST,
		UserID: userID,
	}

	h.taskUC.ScheduleTasks(r.Context(), expression)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": exprID})
}

func (h *Handler) ExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	expressions, err := h.exprUC.GetAll(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions})
}

func (h *Handler) ExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"Missing expression ID"}`, http.StatusBadRequest)
		return
	}

	exprID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	expr, err := h.exprUC.GetByID(r.Context(), userID, exprID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expr)
}

func (h *Handler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	task, err := h.taskUC.GetNextTask(r.Context())
	if err != nil || task.ID == "" {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (h *Handler) PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	userID := 1 //todo to give real value

	if err := h.taskUC.SubmitTaskResult(r.Context(), userID, req.ID, req.Result); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "result accepted"})
}

func (h *Handler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			log.Printf("missing token, error: %v\n", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
			return
		}
		claims, err := h.jwtUtil.ValidateToken(cookie.Value)
		if err != nil {
			log.Printf("invalid token, error: %v\n", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
			return
		}
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, `{"error":"Invalid token claims"}`, http.StatusForbidden)
			return
		}
		userID := int(userIDFloat)

		ctx := context.WithValue(r.Context(), "userID", userID)
		log.Println("AuthMiddleware, userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func convertASTNode(src *calculation.ASTNode) *entity.ASTNode {
	if src == nil {
		return nil
	}
	return &entity.ASTNode{
		IsLeaf:        src.IsLeaf,
		Value:         src.Value,
		Operator:      src.Operator,
		Left:          convertASTNode(src.Left),
		Right:         convertASTNode(src.Right),
		TaskScheduled: src.TaskScheduled,
	}
}
