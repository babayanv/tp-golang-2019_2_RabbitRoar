package http

import (
	"github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/http_utils"
	"github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/models"
	"github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/pack"
	"github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/user"
	"github.com/labstack/echo/v4"
	"github.com/mailru/easyjson"
	"github.com/microcosm-cc/bluemonday"
	"github.com/xeipuuv/gojsonschema"
	"net/http"
	"strconv"
)

type handler struct {
	packUseCase pack.UseCase
	userUseCase user.UseCase
	packSchema  *gojsonschema.Schema
	sanitizer   *bluemonday.Policy
}

func NewPackHandler(
	e *echo.Echo,
	packUseCase pack.UseCase,
	userUseCase user.UseCase,
	authMiddleware echo.MiddlewareFunc,
	csrfMiddleware echo.MiddlewareFunc,
	packSchema *gojsonschema.Schema,
) {
	handler := handler{
		packUseCase: packUseCase,
		userUseCase: userUseCase,
		packSchema:  packSchema,
		sanitizer:   bluemonday.UGCPolicy(),
	}

	group := e.Group("/pack")
	group.POST("", authMiddleware(csrfMiddleware(handler.create)))
	group.DELETE("/:id", authMiddleware(handler.delete))
	group.GET("", authMiddleware(handler.list))
	group.GET("/offline", authMiddleware(handler.offline))
	group.GET("/offline/author", authMiddleware(handler.offlineAuthor))
	group.GET("/offline/public", handler.offlinePublic)
	group.GET("/author", authMiddleware(handler.listAuthor))
	group.GET("/:id", handler.byID)
}

//TODO: make me more safe (or contract that DB has valid form)
func (h *handler) sanitizeQuestions(p interface{}) {
	if p == nil {
		return
	}
	themeSlice := p.([]interface{})
	for _, theme := range themeSlice {
		theme := theme.(map[string]interface{})
		theme["name"] = h.sanitizer.Sanitize(theme["name"].(string))
		questionSlice := theme["questions"].([]interface{})
		for _, question := range questionSlice {
			question := question.(map[string]interface{})
			question["text"] = h.sanitizer.Sanitize(question["text"].(string))
			question["answer"] = h.sanitizer.Sanitize(question["answer"].(string))
		}
	}
}

func (h *handler) sanitize(p models.Pack) models.Pack {
	p.Name = h.sanitizer.Sanitize(p.Name)
	p.Description = h.sanitizer.Sanitize(p.Description)
	p.Tags = h.sanitizer.Sanitize(p.Tags)
	h.sanitizeQuestions(p.Questions)
	return p
}

func (h *handler) sanitizeSlice(p []models.Pack) []models.Pack {
	for i := 0; i < len(p); i++ {
		p[i] = h.sanitize(p[i])
	}
	return p
}

func (h *handler) create(ctx echo.Context) error {
	var p models.Pack
	err := ctx.Bind(&p)
	if err != nil {
		return err
	}

	loader := gojsonschema.NewGoLoader(p.Questions)
	res, err := h.packSchema.Validate(loader)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"error parsing pack",
		)
	}

	if !res.Valid() {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			http_utils.ExtractErrors(res.Errors()),
		)
	}

	caller := ctx.Get("user").(*models.User)
	if err := h.packUseCase.Create(&p, *caller); err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "error creating pack",
			Internal: err,
		}
	}
	_, _, err = easyjson.MarshalToHTTPResponseWriter(h.sanitize(p), ctx.Response())
	return err
}

func (h *handler) offline(ctx echo.Context) error {
	caller := ctx.Get("user").(*models.User)

	ids, err := h.packUseCase.FetchOffline(*caller)
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "error fetching offline",
			Internal: err,
		}
	}

	return ctx.JSON(http.StatusOK, ids)
}

func (h *handler) offlinePublic(ctx echo.Context) error {
	ids, err := h.packUseCase.FetchOfflinePublic()
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "error fetching offline public packs",
			Internal: err,
		}
	}

	return ctx.JSON(http.StatusOK, ids)
}

func (h *handler) offlineAuthor(ctx echo.Context) error {
	caller := ctx.Get("user").(*models.User)

	ids, err := h.packUseCase.FetchOfflineAuthor(*caller)
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "error fetching offline user created packs",
			Internal: err,
		}
	}

	return ctx.JSON(http.StatusOK, ids)
}

func (h *handler) list(ctx echo.Context) error {
	page := http_utils.GetIntParam(ctx, 0)

	packs, err := h.packUseCase.FetchOrderedByRating(true, page, 20)
	if err != nil {
		return &echo.HTTPError{
			Code:     500,
			Message:  "error fetching packs",
			Internal: err,
		}
	}

	return ctx.JSON(http.StatusOK, h.sanitizeSlice(packs))
}

func (h *handler) listAuthor(ctx echo.Context) error {
	caller := ctx.Get("user").(*models.User)
	page := http_utils.GetIntParam(ctx, 0)
	packs, err := h.packUseCase.FetchByAuthor(*caller, true, page, 20)
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "error fetching packs by author",
			Internal: err,
		}
	}

	return ctx.JSON(http.StatusOK, h.sanitizeSlice(packs))
}

func (h *handler) delete(ctx echo.Context) error {
	ID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusBadRequest,
			Message:  "invalid pack id",
			Internal: err,
		}
	}

	p, err := h.packUseCase.GetByID(ID)
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusNotFound,
			Message:  "no pack with such id",
			Internal: err,
		}
	}

	caller := ctx.Get("user").(*models.User)
	if p.Author != caller.ID {
		return echo.NewHTTPError(http.StatusForbidden, "you can delete only own packs")
	}

	if err := h.packUseCase.Delete(ID); err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "error removing pack",
			Internal: err,
		}
	}

	return ctx.NoContent(http.StatusOK)
}

func (h *handler) byID(ctx echo.Context) error {
	ID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusBadRequest,
			Message:  "invalid pack id",
			Internal: err,
		}
	}

	p, err := h.packUseCase.GetByID(ID)
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusNotFound,
			Message:  "no pack with such ID",
			Internal: err,
		}
	}

	if p.Offline {
		return ctx.JSON(http.StatusOK, h.sanitize(*p))
	}

	sessionID, err := ctx.Cookie("SessionID")
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized users can see only offline packs")
	}

	caller, err := h.userUseCase.GetBySessionID(sessionID.Value)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "bad SessionID, invalid or expired")
	}

	if p.Author == caller.ID {
		return ctx.JSON(http.StatusOK, h.sanitize(*p))
	}

	if h.packUseCase.Played(p.ID, caller.ID) {
		return ctx.JSON(http.StatusOK, h.sanitize(*p))
	}

	return echo.NewHTTPError(http.StatusForbidden, "you can view only own, played, created packs")
}
