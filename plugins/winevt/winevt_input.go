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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mozilla-services/heka/message"
	"github.com/mozilla-services/heka/pipeline"
)


type WinEvtInputConfig struct {
	Splitter string
	// From https://msdn.microsoft.com/en-us/library/windows/desktop/aa385487%28v=vs.85%29.aspx
	// The name of the Admin or Operational channel that contains the events that you want to subscribe to.
	ChannelPath string // `toml:"channel_path"`
	// A query that specifies the types of events that you want the subscription service to return.
	Query string
}

type WinEvtInput struct {
	pConfig            *pipeline.PipelineConfig
	config             *WinEvtInputConfig

	name               string
	checkpointFilename string
	bookmark           string
}

func (wi *WinEvtInput) ConfigStruct() interface{} {
	return &WinEvtInputConfig{
		Splitter:    "NullSplitter",
		ChannelPath: "",
		Query:       "*",
	}
}

func (wi *WinEvtInput) writeCheckpoint() (err error) {
	checkpointFile, file_err := os.OpenFile(wi.checkpointFilename,
		os.O_WRONLY|os.O_SYNC|os.O_CREATE|os.O_TRUNC, 0640)
	if file_err != nil {
		return fmt.Errorf("Error opening winevt checkpoint %s: %s", wi.checkpointFilename, file_err.Error())
	}
	defer checkpointFile.Close()
	if _, file_err = checkpointFile.WriteString(wi.bookmark); file_err != nil {
		return fmt.Errorf("Error writing winevt checkpoint %s: %s", wi.checkpointFilename, file_err.Error())
	}
	return nil
}

func (wi *WinEvtInput) readCheckpoint() (err error) {
	checkpointFile, file_err := os.Open(wi.checkpointFilename)
	if file_err != nil {
		return fmt.Errorf("Error opening winevt checkpoint %s: %s", wi.checkpointFilename, file_err.Error())
	}
	contents := bytes.NewBuffer(nil)
	defer checkpointFile.Close()
	io.Copy(contents, checkpointFile)
	wi.bookmark = contents.String()
	return
}

func (wi *WinEvtInput) SetPipelineConfig(pConfig *pipeline.PipelineConfig) {
	wi.pConfig = pConfig
}

func (wi *WinEvtInput) SetName(name string) {
    wi.name = name
}

func (wi *WinEvtInput) Init(config interface{}) (err error) {
	wi.config = config.(*WinEvtInputConfig)
	wi.checkpointFilename = wi.pConfig.Globals.PrependBaseDir(filepath.Join("winevt",
		fmt.Sprintf("%s.offset.bin", wi.name)))
	wi.readCheckpoint()
	return nil
}

func (wi *WinEvtInput) Run(ir pipeline.InputRunner, h pipeline.PluginHelper) (err error) {
	sRunner := ir.NewSplitterRunner("")

	if !sRunner.UseMsgBytes() {
		packDec := func(pack *pipeline.PipelinePack) {
			pack.Message.SetType("winevt")
			pack.Message.SetLogger(wi.name)
			field, err := message.NewField("channel_path", wi.config.ChannelPath, "")
			if err != nil {
				sRunner.LogError(
					fmt.Errorf("can't add 'channel_path' field: %s", err.Error()))
			} else {
				pack.Message.AddField(field)
			}
		}
		sRunner.SetPackDecorator(packDec)
    }

	return nil
}

func (wi *WinEvtInput) Stop() {
}

func init() {
	pipeline.RegisterPlugin("WinEvtInput", func() interface{} {
		return new(WinEvtInput)
	})
}
