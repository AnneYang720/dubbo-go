/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"encoding/json"
	"net/url"
	"strconv"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

import (
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
)

func TestDefaultAuthenticator_Authenticate(t *testing.T) {
	secret := "dubbo-sk"
	access := "dubbo-ak"
	testurl, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider?interface=com.ikurento.user.UserProvider&group=gg&version=2.6.0")
	testurl.SetParam(constant.AccessKeyIDKey, access)
	testurl.SetParam(constant.SecretAccessKeyKey, secret)
	parmas := []interface{}{"OK", struct {
		Name string
		ID   int64
	}{"YUYU", 1}}

	requestTime := strconv.Itoa(int(time.Now().Unix() * 1000))
	
	content, _ := json.Marshal(parmas)
	signature := doSign(content, secret)

	authenticator = &defaultAuthenticator{}

	invcation := invocation.NewRPCInvocation("test", parmas, map[string]interface{}{
		constant.RequestSignatureKey: signature,
		constant.Consumer:            "test",
		constant.RequestTimestampKey: requestTime,
		constant.AKKey:               access,
		"content":				      string(content),
	})
	err := authenticator.Authenticate(invcation, testurl)
	assert.Nil(t, err)

	// modify the params
	content, _ = json.Marshal(parmas[:1])
	invcation = invocation.NewRPCInvocation("test", parmas[:1], map[string]interface{}{
		constant.RequestSignatureKey: signature,
		constant.Consumer:            "test",
		constant.RequestTimestampKey: requestTime,
		constant.AKKey:               access,
		"content":				      string(content),
	})
	err = authenticator.Authenticate(invcation, testurl)
	assert.NotNil(t, err)
}

func TestDefaultAuthenticator_Sign(t *testing.T) {
	authenticator = &defaultAuthenticator{}
	testurl, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider?application=test&interface=com.ikurento.user.UserProvider&group=gg&version=2.6.0")
	testurl.SetParam(constant.AccessKeyIDKey, "akey")
	testurl.SetParam(constant.SecretAccessKeyKey, "skey")
	testurl.SetParam(constant.ParameterSignatureEnableKey, "false")
	inv := invocation.NewRPCInvocation("test", []interface{}{"OK"}, nil)
	_ = authenticator.Sign(inv, testurl)
	assert.NotEqual(t, inv.GetAttachmentWithDefaultValue(constant.Consumer, ""), "")
	assert.NotEqual(t, inv.GetAttachmentWithDefaultValue(constant.RequestTimestampKey, ""), "")
	assert.Equal(t, inv.GetAttachmentWithDefaultValue(constant.AKKey, ""), "akey")
	assert.Equal(t, inv.GetAttachmentWithDefaultValue(constant.SKKey, ""), "skey")
}

func Test_getAccessKeyPairSuccess(t *testing.T) {
	testurl := common.NewURLWithOptions(
		common.WithParams(url.Values{}),
		common.WithParamsValue(constant.SecretAccessKeyKey, "skey"),
		common.WithParamsValue(constant.AccessKeyIDKey, "akey"))
	invcation := invocation.NewRPCInvocation("MethodName", []interface{}{"OK"}, nil)
	_, e := getAccessKeyPair(invcation, testurl)
	assert.Nil(t, e)
}

func Test_getAccessKeyPairFailed(t *testing.T) {
	defer func() {
		e := recover()
		assert.NotNil(t, e)
	}()
	testurl := common.NewURLWithOptions(
		common.WithParams(url.Values{}),
		common.WithParamsValue(constant.AccessKeyIDKey, "akey"))
	invcation := invocation.NewRPCInvocation("MethodName", []interface{}{"OK"}, nil)
	_, e := getAccessKeyPair(invcation, testurl)
	assert.NotNil(t, e)
	testurl = common.NewURLWithOptions(
		common.WithParams(url.Values{}),
		common.WithParamsValue(constant.SecretAccessKeyKey, "skey"),
		common.WithParamsValue(constant.AccessKeyIDKey, "akey"), common.WithParamsValue(constant.AccessKeyStorageKey, "dubbo"))
	_, e = getAccessKeyPair(invcation, testurl)
	assert.NoError(t, e)
}
