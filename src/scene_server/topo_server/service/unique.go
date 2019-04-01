/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"strconv"

	"configcenter/src/auth/meta"
	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/scene_server/topo_server/core/types"
)

// CreateObjectUnique create a new object unique
func (s *Service) CreateObjectUnique(params types.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {
	request := &metadata.CreateUniqueRequest{}

	if err := data.MarshalJSONInto(request); err != nil {
		blog.Errorf("[CreateObjectUnique] unmarshal error: %v, data: %#v", err, data)
		return nil, params.Err.New(common.CCErrCommParamsInvalid, err.Error())
	}

	objectID := pathParams(common.BKObjIDField)

	// auth: check authorization
	if err := s.AuthManager.AuthorizeModelUniqueResourceCreate(params.Context, params.Header, objectID); err != nil {
		blog.Errorf("create model unique failed, authorization failed, modelID: %s, err: %+v", objectID, err)
		return nil, params.Err.New(common.CCErrCommAuthNotHavePermission, err.Error())
	}

	id, err := s.Core.UniqueOperation().Create(params, objectID, request)
	if err != nil {
		blog.Errorf("[CreateObjectUnique] create for [%s] failed: %v, raw: %#v", objectID, err, data)
		return nil, err
	}

	uniqueID := id.ID

	// auth: register model unique
	if err := s.AuthManager.RegisterModuleUniqueByID(params.Context, params.Header, uniqueID); err != nil {
		blog.Errorf("register model unique to iam failed, uniqueID: %d, err: %+v", uniqueID, err)
		return nil, params.Err.New(common.CCErrCommUnRegistResourceToIAMFailed, err.Error())
	}

	return id, nil
}

// UpdateObjectUnique update a object unique
func (s *Service) UpdateObjectUnique(params types.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {
	request := &metadata.UpdateUniqueRequest{}

	if err := data.MarshalJSONInto(request); err != nil {
		blog.Errorf("[UpdateObjectUnique] unmarshal error: %v, data: %#v", err, data)
		return nil, params.Err.New(common.CCErrCommParamsInvalid, err.Error())
	}

	objectID := pathParams(common.BKObjIDField)
	id, err := strconv.ParseUint(pathParams("id"), 10, 64)
	if err != nil {
		return nil, params.Err.Errorf(common.CCErrCommParamsInvalid, "id")
	}

	data.Remove(metadata.BKMetadata)

	// auth: check authorization
	if err := s.AuthManager.AuthorizeByUniqueID(params.Context, params.Header, meta.Update, int64(id)); err != nil {
		blog.Errorf("update model unique failed, authorization failed, unique ID: %d, err: %+v", id, err)
		return nil, params.Err.New(common.CCErrCommAuthNotHavePermission, err.Error())
	}
	err = s.Core.UniqueOperation().Update(params, objectID, id, request)
	if err != nil {
		blog.Errorf("[UpdateObjectUnique] update for [%s](%d) failed: %v, raw: %#v", objectID, id, err, data)
		return nil, err
	}
	// auth: update registered model unique
	if err := s.AuthManager.UpdateRegisteredModelUniqueByID(params.Context, params.Header, int64(id)); err != nil {
		blog.Errorf("update register model unique to iam failed, uniqueID: %d, err: %+v", id, err)
		return nil, params.Err.New(common.CCErrCommRegistResourceToIAMFailed, err.Error())
	}
	return nil, nil
}

// DeleteObjectUnique delete a object unique
func (s *Service) DeleteObjectUnique(params types.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {
	objectID := pathParams(common.BKObjIDField)
	id, err := strconv.ParseUint(pathParams("id"), 10, 64)
	if err != nil {
		return nil, params.Err.Errorf(common.CCErrCommParamsInvalid, "id")
	}
	data.Remove(metadata.BKMetadata)

	uniques, err := s.Core.UniqueOperation().Search(params, objectID)
	if err != nil {
		return nil, err
	}

	if len(uniques) <= 1 {
		blog.Errorf("[DeleteObjectUnique][%s] unique should have more than one", objectID)
		return nil, params.Err.Error(common.CCErrTopoObjectUniqueShouldHaveMoreThanOne)
	}

	// auth: check authorization
	if err := s.AuthManager.AuthorizeByUniqueID(params.Context, params.Header, meta.Update, int64(id)); err != nil {
		blog.Errorf("delete model unique failed, authorization failed, unique ID: %d, err: %+v", id, err)
		return nil, params.Err.New(common.CCErrCommAuthNotHavePermission, err.Error())
	}

	err = s.Core.UniqueOperation().Delete(params, objectID, id)
	if err != nil {
		blog.Errorf("[DeleteObjectUnique] delete [%s](%d) failed: %v", objectID, id, err)
		return nil, err
	}

	// auth: update registered model unique
	if err := s.AuthManager.DeregisterModelUniqueByID(params.Context, params.Header, int64(id)); err != nil {
		blog.Errorf("deregister model unique from iam failed, uniqueID: %d, err: %+v", id, err)
		return nil, params.Err.New(common.CCErrCommUnRegistResourceToIAMFailed, err.Error())
	}

	return nil, nil
}

// SearchObjectUnique search object uniques
func (s *Service) SearchObjectUnique(params types.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {
	objectID := pathParams(common.BKObjIDField)
	uniques, err := s.Core.UniqueOperation().Search(params, objectID)
	if err != nil {
		blog.Errorf("[SearchObjectUnique] search for [%s] failed: %v", objectID, err)
		return nil, err
	}

	// auth: check authorization
	if err := s.AuthManager.AuthorizeByUnique(params.Context, params.Header, meta.Update, uniques...); err != nil {
		blog.Errorf("update model unique failed, authorization failed, unique: %+v, err: %+v", uniques, err)
		return nil, params.Err.New(common.CCErrCommAuthNotHavePermission, err.Error())
	}

	return uniques, nil
}
