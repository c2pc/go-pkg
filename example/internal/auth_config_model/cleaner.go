package auth_config_model

import (
	"encoding/json"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/gin-gonic/gin/binding"
)

type CleanConfig struct {
	ch chan int
}

type CleanCfg struct {
	HistoryAccDays int `json:"history_acc_days" binding:"required,gt=10"`
	MaxRows        int `json:"max_rows" binding:"required,gt=10"`
}

func NewCleanConfig() CleanConfig {
	return CleanConfig{
		ch: make(chan int, 10),
	}
}

func (c *CleanConfig) Check(data []byte) error {
	var r CleanCfg
	if err := binding.JSON.BindBody(data, &r); err != nil {
		return apperr.ErrValidation.WithError(err)
	}

	return nil
}

func (c *CleanConfig) AfterUpdate(data []byte) error {
	return nil
}

func (c *CleanConfig) Init() ([]byte, error) {
	cfg := CleanCfg{
		HistoryAccDays: 365,
		MaxRows:        10000,
	}

	return json.Marshal(cfg)
}

type CleanCfgTransformer struct {
	HistoryAccDays int `json:"history_acc_days"`
	MaxRows        int `json:"max_rows"`
}

func (c *CleanConfig) Transform(data []byte) (any, error) {
	var cfg CleanCfg
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return CleanCfgTransformer(cfg), nil
}
