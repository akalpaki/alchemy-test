package spacecraft

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/akalpaki/alchemy-test/pkg/web"
)

func Routes(repo *Repository, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/spacecrafts", handleCreate(repo, logger))
	mux.HandleFunc("PUT /v1/spacecrafts/{id}", handleUpdate(repo, logger))
	mux.HandleFunc("DELETE /v1/spacecrafts/{id}", handleDelete(repo, logger))
	mux.HandleFunc("GET /v1/spacecrafts/{id}", handleGetByID(repo, logger))
	mux.HandleFunc("GET /v1/spacecrafts/", handleGet(repo, logger))

	return mux
}

func handleCreate(repo *Repository, logger *slog.Logger) http.HandlerFunc {
	type response struct {
		success string `json:"success"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqData, err := web.ReadJSON[SpacecraftRequest](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid or malformed json", err)
		}

		err = repo.Create(ctx, reqData)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to create entry", err)
		}

		if err := web.WriteJSON(w, r, http.StatusCreated, response{success: "true"}); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to provide response", err)
		}
	}
}

func handleUpdate(repo *Repository, logger *slog.Logger) http.HandlerFunc {
	type response struct {
		success string `json:"success"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid spaceship id", err)
		}

		reqData, err := web.ReadJSON[SpacecraftRequest](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid or malformed json", err)
		}

		err = repo.Update(ctx, id, reqData)
		if err != nil {
			switch err {
			case errNotFound:
				web.ErrorResponse(logger, w, r, http.StatusNotFound, "entry not found", err)
			default:
				web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to update entry", err)
			}
		}

		if err := web.WriteJSON(w, r, http.StatusCreated, response{success: "true"}); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to provide response", err)
		}
	}
}

func handleDelete(repo *Repository, logger *slog.Logger) http.HandlerFunc {
	type response struct {
		success string `json:"success"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid spaceship id", err)
		}

		err = repo.Delete(ctx, id)
		if err != nil {
			switch err {
			case errNotFound:
				web.ErrorResponse(logger, w, r, http.StatusNotFound, "entry not found", err)
			default:
				web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to delete entry", err)
			}
		}

		if err := web.WriteJSON(w, r, http.StatusCreated, response{success: "true"}); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to provide response", err)
		}
	}
}

func handleGetByID(repo *Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid id", err)
		}

		spacecraft, err := repo.GetByID(ctx, id)
		if err != nil {
			switch err {
			case errNotFound:
				web.ErrorResponse(logger, w, r, http.StatusNotFound, "entry not found", err)
			default:
				web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to retrieve entry", err)
			}
		}

		if err := web.WriteJSON(w, r, http.StatusOK, spacecraft); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
		}
	}
}

// filter by name, class, status
func handleGet(repo *Repository, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		filters := r.URL.Query()

		spacecrafts, err := repo.Get(ctx, filters)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to retrieve spaceships", err)
		}

		if err := web.WriteJSON(w, r, http.StatusOK, spacecrafts); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
		}
	}
}
