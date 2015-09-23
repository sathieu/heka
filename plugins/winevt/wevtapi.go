/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2014-2015
# the Initial Developer. All Rights Reserved.
#
# Contributor(s):
#   Mathieu Parent (math.parent@gmail.com)
#
# ***** END LICENSE BLOCK *****/
package winevt

import (
	"syscall"
	"unsafe"
)

var (
	modwevtapi = syscall.NewLazyDLL("wevtapi.dll")

	procEvtSubscribe = modwevtapi.NewProc("EvtSubscribeW")
	procEvtQuery =     modwevtapi.NewProc("EvtQueryW")
	procEvtClose =     modwevtapi.NewProc("EvtCloseW")
)

// EVT_SUBSCRIBE_FLAGS
type EVT_SUBSCRIBE_FLAGS int32
const (
  EvtSubscribeToFutureEvents       EVT_SUBSCRIBE_FLAGS = 1
  EvtSubscribeStartAtOldestRecord  EVT_SUBSCRIBE_FLAGS = 2
  EvtSubscribeStartAfterBookmark   EVT_SUBSCRIBE_FLAGS = 3
  EvtSubscribeOriginMask           EVT_SUBSCRIBE_FLAGS = 0x3
  EvtSubscribeTolerateQueryErrors  EVT_SUBSCRIBE_FLAGS = 0x1000
  EvtSubscribeStrict               EVT_SUBSCRIBE_FLAGS = 0x10000
)

// EVT_SUBSCRIBE_NOTIFY_ACTION
type EVT_SUBSCRIBE_NOTIFY_ACTION int32
const (
  EvtSubscribeActionError   EVT_SUBSCRIBE_NOTIFY_ACTION = 0
  EvtSubscribeActionDeliver EVT_SUBSCRIBE_NOTIFY_ACTION = 1
)

// Simple types
type DWORD uint32
type HANDLE uintptr
type EVT_HANDLE HANDLE
type PVOID unsafe.Pointer

// Callbacks
type EVT_SUBSCRIBE_CALLBACK func(action EVT_SUBSCRIBE_NOTIFY_ACTION, userContext PVOID, event EVT_HANDLE) (DWORD)

func EvtSubscribe(session EVT_HANDLE, signalEvent HANDLE, channelPath, query string,
	bookmark EVT_HANDLE, context PVOID, callback EVT_SUBSCRIBE_CALLBACK, flags EVT_SUBSCRIBE_FLAGS) (EVT_HANDLE, error) {
	ret, _, callErr := procEvtSubscribe.Call(
		uintptr(session),
		uintptr(signalEvent),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(channelPath))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(query))),
		uintptr(bookmark),
		uintptr(context),
		uintptr(syscall.NewCallback(callback)),
		uintptr(flags))
    if ret == 0 {
        return 0, callErr
    }
    return EVT_HANDLE(ret), nil
}

func EvtQuery(session EVT_HANDLE, path, query string, flags DWORD) (EVT_HANDLE, error) {
	ret, _, callErr := procEvtSubscribe.Call(
		uintptr(session),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(query))),
		uintptr(flags))
    if ret == 0 {
        return 0, callErr
    }
    return EVT_HANDLE(ret), nil
}

func EvtClose(object EVT_HANDLE) (bool, error) {
    ret, _, callErr := procEvtClose.Call(
        uintptr(object))
    return ret != 0, callErr
}
