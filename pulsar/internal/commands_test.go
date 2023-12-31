// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertStringMap(t *testing.T) {
	m := make(map[string]string)
	m["a"] = "1"
	m["b"] = "2"

	pbm := ConvertFromStringMap(m)

	assert.Equal(t, 2, len(pbm))

	m2 := ConvertToStringMap(pbm)
	assert.Equal(t, 2, len(m2))
	assert.Equal(t, "1", m2["a"])
	assert.Equal(t, "2", m2["b"])
}

func TestReadMessageMetadata(t *testing.T) {
	// read old style message (not batched)
	reader := NewMessageReaderFromArray(rawCompatSingleMessage)
	meta, err := reader.ReadMessageMetadata()
	if err != nil {
		t.Fatal(err)
	}

	props := meta.GetProperties()
	assert.Equal(t, len(props), 2)
	assert.Equal(t, "a", props[0].GetKey())
	assert.Equal(t, "1", props[0].GetValue())
	assert.Equal(t, "b", props[1].GetKey())
	assert.Equal(t, "2", props[1].GetValue())

	// read message with batch of 1
	reader = NewMessageReaderFromArray(rawBatchMessage1)
	meta, err = reader.ReadMessageMetadata()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, int(meta.GetNumMessagesInBatch()))

	// read message with batch of 10
	reader = NewMessageReaderFromArray(rawBatchMessage10)
	meta, err = reader.ReadMessageMetadata()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10, int(meta.GetNumMessagesInBatch()))
}

func TestReadBrokerEntryMetadata(t *testing.T) {
	// read old style message (not batched)
	reader := NewMessageReaderFromArray(brokerEntryMeta)
	meta, err := reader.ReadBrokerMetadata()
	if err != nil {
		t.Fatal(err)
	}
	var expectedBrokerTimestamp uint64 = 1646983036054
	assert.Equal(t, expectedBrokerTimestamp, *meta.BrokerTimestamp)
	var expectedIndex uint64 = 5
	assert.Equal(t, expectedIndex, *meta.Index)
}

func TestReadMessageOldFormat(t *testing.T) {
	reader := NewMessageReaderFromArray(rawCompatSingleMessage)
	_, err := reader.ReadMessageMetadata()
	if err != nil {
		t.Fatal(err)
	}

	ssm, payload, err := reader.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	// old message format does not have a single message metadata
	assert.Equal(t, true, ssm == nil)
	assert.Equal(t, "hello", string(payload))

	_, _, err = reader.ReadMessage()
	assert.Equal(t, ErrEOM, err)
}

func TestReadMessagesBatchSize1(t *testing.T) {
	reader := NewMessageReaderFromArray(rawBatchMessage1)
	meta, err := reader.ReadMessageMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, int(meta.GetNumMessagesInBatch()))
	for i := 0; i < int(meta.GetNumMessagesInBatch()); i++ {
		ssm, payload, err := reader.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, ssm != nil)
		assert.Equal(t, "hello", string(payload))
	}

	_, _, err = reader.ReadMessage()
	assert.Equal(t, ErrEOM, err)
}

func TestReadMessagesBatchSize10(t *testing.T) {
	reader := NewMessageReaderFromArray(rawBatchMessage10)
	meta, err := reader.ReadMessageMetadata()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10, int(meta.GetNumMessagesInBatch()))
	for i := 0; i < int(meta.GetNumMessagesInBatch()); i++ {
		ssm, payload, err := reader.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, ssm != nil)
		assert.Equal(t, "hello", string(payload))
	}

	_, _, err = reader.ReadMessage()
	assert.Equal(t, ErrEOM, err)
}

// Raw single message in old format
// metadata properties:<key:"a" value:"1" > properties:<key:"b" value:"2" >
// payload = "hello"
var rawCompatSingleMessage = []byte{
	0x0e, 0x01, 0x08, 0x36, 0xb4, 0x66, 0x00, 0x00,
	0x00, 0x31, 0x0a, 0x0f, 0x73, 0x74, 0x61, 0x6e,
	0x64, 0x61, 0x6c, 0x6f, 0x6e, 0x65, 0x2d, 0x37,
	0x34, 0x2d, 0x30, 0x10, 0x00, 0x18, 0xac, 0xef,
	0xe8, 0xa0, 0xe2, 0x2d, 0x22, 0x06, 0x0a, 0x01,
	0x61, 0x12, 0x01, 0x31, 0x22, 0x06, 0x0a, 0x01,
	0x62, 0x12, 0x01, 0x32, 0x48, 0x05, 0x60, 0x05,
	0x82, 0x01, 0x00, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
}

// Message with batch of 1
// singe message metadata properties:<key:"a" value:"1" > properties:<key:"b" value:"2" >
// payload = "hello"
var rawBatchMessage1 = []byte{
	0x0e, 0x01, 0x1f, 0x80, 0x09, 0x68, 0x00, 0x00,
	0x00, 0x1f, 0x0a, 0x0f, 0x73, 0x74, 0x61, 0x6e,
	0x64, 0x61, 0x6c, 0x6f, 0x6e, 0x65, 0x2d, 0x37,
	0x34, 0x2d, 0x31, 0x10, 0x00, 0x18, 0xdb, 0x80,
	0xf4, 0xa0, 0xe2, 0x2d, 0x58, 0x01, 0x82, 0x01,
	0x00, 0x00, 0x00, 0x00, 0x16, 0x0a, 0x06, 0x0a,
	0x01, 0x61, 0x12, 0x01, 0x31, 0x0a, 0x06, 0x0a,
	0x01, 0x62, 0x12, 0x01, 0x32, 0x18, 0x05, 0x28,
	0x05, 0x40, 0x00, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
}

// Message with batch of 10
// singe message metadata properties:<key:"a" value:"1" > properties:<key:"b" value:"2" >
// payload = "hello"
var rawBatchMessage10 = []byte{
	0x0e, 0x01, 0x7b, 0x28, 0x8c, 0x08,
	0x00, 0x00, 0x00, 0x1f, 0x0a, 0x0f, 0x73, 0x74,
	0x61, 0x6e, 0x64, 0x61, 0x6c, 0x6f, 0x6e, 0x65,
	0x2d, 0x37, 0x34, 0x2d, 0x32, 0x10, 0x00, 0x18,
	0xd0, 0xc2, 0xfa, 0xa0, 0xe2, 0x2d, 0x58, 0x0a,
	0x82, 0x01, 0x00, 0x00, 0x00, 0x00, 0x16, 0x0a,
	0x06, 0x0a, 0x01, 0x61, 0x12, 0x01, 0x31, 0x0a,
	0x06, 0x0a, 0x01, 0x62, 0x12, 0x01, 0x32, 0x18,
	0x05, 0x28, 0x05, 0x40, 0x00, 0x68, 0x65, 0x6c,
	0x6c, 0x6f, 0x00, 0x00, 0x00, 0x16, 0x0a, 0x06,
	0x0a, 0x01, 0x61, 0x12, 0x01, 0x31, 0x0a, 0x06,
	0x0a, 0x01, 0x62, 0x12, 0x01, 0x32, 0x18, 0x05,
	0x28, 0x05, 0x40, 0x01, 0x68, 0x65, 0x6c, 0x6c,
	0x6f, 0x00, 0x00, 0x00, 0x16, 0x0a, 0x06, 0x0a,
	0x01, 0x61, 0x12, 0x01, 0x31, 0x0a, 0x06, 0x0a,
	0x01, 0x62, 0x12, 0x01, 0x32, 0x18, 0x05, 0x28,
	0x05, 0x40, 0x02, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
	0x00, 0x00, 0x00, 0x16, 0x0a, 0x06, 0x0a, 0x01,
	0x61, 0x12, 0x01, 0x31, 0x0a, 0x06, 0x0a, 0x01,
	0x62, 0x12, 0x01, 0x32, 0x18, 0x05, 0x28, 0x05,
	0x40, 0x03, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x00,
	0x00, 0x00, 0x16, 0x0a, 0x06, 0x0a, 0x01, 0x61,
	0x12, 0x01, 0x31, 0x0a, 0x06, 0x0a, 0x01, 0x62,
	0x12, 0x01, 0x32, 0x18, 0x05, 0x28, 0x05, 0x40,
	0x04, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x00, 0x00,
	0x00, 0x16, 0x0a, 0x06, 0x0a, 0x01, 0x61, 0x12,
	0x01, 0x31, 0x0a, 0x06, 0x0a, 0x01, 0x62, 0x12,
	0x01, 0x32, 0x18, 0x05, 0x28, 0x05, 0x40, 0x05,
	0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x00, 0x00, 0x00,
	0x16, 0x0a, 0x06, 0x0a, 0x01, 0x61, 0x12, 0x01,
	0x31, 0x0a, 0x06, 0x0a, 0x01, 0x62, 0x12, 0x01,
	0x32, 0x18, 0x05, 0x28, 0x05, 0x40, 0x06, 0x68,
	0x65, 0x6c, 0x6c, 0x6f, 0x00, 0x00, 0x00, 0x16,
	0x0a, 0x06, 0x0a, 0x01, 0x61, 0x12, 0x01, 0x31,
	0x0a, 0x06, 0x0a, 0x01, 0x62, 0x12, 0x01, 0x32,
	0x18, 0x05, 0x28, 0x05, 0x40, 0x07, 0x68, 0x65,
	0x6c, 0x6c, 0x6f, 0x00, 0x00, 0x00, 0x16, 0x0a,
	0x06, 0x0a, 0x01, 0x61, 0x12, 0x01, 0x31, 0x0a,
	0x06, 0x0a, 0x01, 0x62, 0x12, 0x01, 0x32, 0x18,
	0x05, 0x28, 0x05, 0x40, 0x08, 0x68, 0x65, 0x6c,
	0x6c, 0x6f, 0x00, 0x00, 0x00, 0x16, 0x0a, 0x06,
	0x0a, 0x01, 0x61, 0x12, 0x01, 0x31, 0x0a, 0x06,
	0x0a, 0x01, 0x62, 0x12, 0x01, 0x32, 0x18, 0x05,
	0x28, 0x05, 0x40, 0x09, 0x68, 0x65, 0x6c, 0x6c,
	0x6f,
}

var brokerEntryMeta = []byte{
	0x0e, 0x02, 0x00, 0x00, 0x00, 0x09, 0x08, 0x96,
	0xf9, 0xda, 0xbe, 0xf7, 0x2f, 0x10, 0x05,
}
