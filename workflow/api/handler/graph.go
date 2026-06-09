package handler

import (
	"net/http"
	"strconv"

	"github.com/engine-go/workflow/dao"
	"github.com/engine-go/workflow/models"
	"github.com/gin-gonic/gin"
)

type GraphHandler struct {
	dao *dao.WfGraphDao
}

func NewGraphHandler() *GraphHandler {
	return &GraphHandler{dao: dao.NewWfGraphDao(nil)}
}

type respEnvelope struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Data any    `json:"data,omitempty"`
}

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, respEnvelope{Code: 0, Data: data})
}

func fail(c *gin.Context, httpStatus int, msg string) {
	c.JSON(httpStatus, respEnvelope{Code: httpStatus, Msg: msg})
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func (h *GraphHandler) List(c *gin.Context) {
	q := &dao.WfGraphQuery{
		GraphID:  c.Query("graph_id"),
		Name:     c.Query("name"),
		Type:     c.Query("type"),
		CreateBy: c.Query("create_by"),
	}
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			st := models.WfGraphStatus(v)
			q.Status = &st
		}
	}
	if s := c.Query("record_id"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			q.RecordID = &v
		}
	}
	q.Limit = 20
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		q.Limit = v
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v >= 0 {
		q.Offset = v
	}

	list, total, err := h.dao.List(c.Request.Context(), q)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, gin.H{"list": list, "total": total, "limit": q.Limit, "offset": q.Offset})
}

func (h *GraphHandler) Get(c *gin.Context) {
	id, okID := parseID(c)
	if !okID {
		return
	}
	g, err := h.dao.GetByID(c.Request.Context(), id)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if g == nil {
		fail(c, http.StatusNotFound, "wf_graph not found")
		return
	}
	ok(c, g)
}

func (h *GraphHandler) Create(c *gin.Context) {
	var g models.WfGraph
	if err := c.ShouldBindJSON(&g); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	g.ID = 0
	if g.Status == 0 {
		g.Status = models.WfGraphStatusDraft
	}
	if g.Version == 0 {
		g.Version = 1
	}
	if err := h.dao.Create(c.Request.Context(), &g); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, g)
}

func (h *GraphHandler) Update(c *gin.Context) {
	id, okID := parseID(c)
	if !okID {
		return
	}
	var g models.WfGraph
	if err := c.ShouldBindJSON(&g); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	g.ID = id
	if err := h.dao.Update(c.Request.Context(), &g); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, gin.H{"id": id})
}

// Detail 兼容前端写死的接口路径：POST /onboard/api/v1/workflow/graph/detail
// 请求体 {graph_id: "..."}；响应 {data:{graph: WfGraph}}，
// 页面会取 data.graph.graph 反序列化成 {nodes:[...]}。
type detailReq struct {
	GraphID string `json:"graph_id"`
}

func (h *GraphHandler) Detail(c *gin.Context) {
	var req detailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.GraphID == "" {
		fail(c, http.StatusBadRequest, "graph_id is required")
		return
	}
	g, err := h.dao.GetByGraphID(c.Request.Context(), req.GraphID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if g == nil {
		fail(c, http.StatusNotFound, "graph_id not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"graph": g}})
}

func (h *GraphHandler) Delete(c *gin.Context) {
	id, okID := parseID(c)
	if !okID {
		return
	}
	updateBy := c.Query("update_by")
	if updateBy == "" {
		updateBy = "system"
	}
	if err := h.dao.Delete(c.Request.Context(), id, updateBy); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, gin.H{"id": id})
}
