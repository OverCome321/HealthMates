// internal/transport/rest/oauth.go
package rest

import (
	"net/http"

	"healthmates/auth-service/utils"
)

// @Summary      Начало OAuth Google
// @Description  Перенаправление на Google для аутентификации
// @Tags         auth
// @Success      302
// @Router       /auth/google [get]
func (h *Handler) handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	state, _ := utils.GenerateState()
	utils.SetStateCookie(w, state)
	url := h.auth.GetGoogleAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// @Summary      Callback OAuth Google
// @Description  Обрабатывает callback от Google и выдаёт JWT
// @Tags         auth
// @Param        code   query     string  true  "Auth Code"
// @Param        state  query     string  true  "CSRF State"
// @Success      302
// @Failure      400 {object} utils.APIError
// @Failure      500 {object} utils.APIError
// @Router       /auth/google/callback [get]
func (h *Handler) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	returned, _ := utils.GetStateFromCookie(r)
	if r.FormValue("state") != returned {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	tok, err := h.auth.HandleGoogleCallback(r.Context(), returned, r.FormValue("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/?token="+tok, http.StatusSeeOther)
}
