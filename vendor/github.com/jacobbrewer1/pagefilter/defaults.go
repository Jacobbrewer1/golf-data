package pagefilter

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	_ "github.com/jacobbrewer1/pagefilter/common"
	"github.com/spf13/viper"
)

//go:embed common/common.yaml
var specFile embed.FS

const (
	baseConfigKey         = "components.parameters"
	configKeyLimitDefault = baseConfigKey + ".limit_param.schema.default"
	configKeyLimitMin     = baseConfigKey + ".limit_param.schema.minimum"
	configKeyLimitMax     = baseConfigKey + ".limit_param.schema.maximum"
)

var specViper = sync.OnceValue(func() viper.Viper {
	file, err := specFile.ReadFile("common/common.yaml")
	if err != nil {
		panic(fmt.Errorf("failed to read spec file: %w", err))
	}

	vip := viper.New()
	vip.SetConfigType("yaml")
	if err := vip.ReadConfig(strings.NewReader(string(file))); err != nil {
		panic(fmt.Errorf("failed to read spec file: %w", err))
	}

	return *vip
})

// limitDefault is the default limit for pagination.
var limitDefault = sync.OnceValue(func() int {
	vip := specViper()
	vip.SetDefault(configKeyLimitDefault, queryLimitDefault)

	// If a config value is set, use it; otherwise, use the default
	limit := vip.GetInt(configKeyLimitDefault)
	if limit < limitMin() {
		return queryLimitDefault
	} else if limit > limitMax() {
		return limitMax()
	}
	return limit
})

// limitMin is the minimum limit for pagination
var limitMin = sync.OnceValue(func() int {
	vip := specViper()
	vip.SetDefault(configKeyLimitMin, queryLimitMin)

	// If a config value is set, use it; otherwise, use the default
	limit := vip.GetInt(configKeyLimitMin)
	if limit <= 0 {
		return queryLimitMin
	}
	return limit
})

// limitMax is the maximum limit for pagination
var limitMax = sync.OnceValue(func() int {
	vip := specViper()
	vip.SetDefault(configKeyLimitMax, queryLimitMax)

	// If a config value is set, use it; otherwise, use the default
	limit := vip.GetInt(configKeyLimitMax)
	if limit <= 0 {
		return queryLimitMax
	}
	return limit
})
