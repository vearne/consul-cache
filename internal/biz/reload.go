package biz

import (
	"context"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/vearne/consul-cache/internal/consts"
	"github.com/vearne/consul-cache/internal/model"
	"github.com/vearne/consul-cache/internal/resource"
	"strconv"
	"time"
)

func ReloadFromRedis(dc string, svc string, expiration time.Duration) (*model.ServiceState, error) {
	var state model.ServiceState
	ctx := context.Background()
	//index:{dc}:{service}
	key := fmt.Sprintf(consts.IndexkeyFormat, dc, svc)
	idxStr, err := resource.StateReadOnlyRedis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			state.Data = make([]consulapi.ServiceEntry, 0)
			return &state, nil
		} else {
			return nil, errors.WithMessage(err, "StateReadOnlyRedis.Get")
		}
	}

	index, err := strconv.Atoi(idxStr)
	if err != nil {
		return nil, errors.WithMessage(err, "strconv.Atoi")
	}
	state.Index = uint64(index)

	key = fmt.Sprintf(consts.DatakeyFormat, dc, svc, idxStr)
	data, err := resource.StateReadOnlyRedis.Get(ctx, key).Result()
	if err != nil {
		return nil, errors.WithMessage(err, "StateReadOnlyRedis.Get")
	}

	state.Data = make([]consulapi.ServiceEntry, 0)
	jsoniter.Unmarshal([]byte(data), &state.Data)

	key = fmt.Sprintf(consts.StatekeyFormat, dc, svc)
	resource.SeviceStateCache.Set(key, &state, expiration)

	return &state, nil
}
