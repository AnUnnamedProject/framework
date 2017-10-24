// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"encoding/json"
	"fmt"
)

// FlashData maintains data across requests.
type FlashData struct {
	Data map[string]string
}

// NewFlash returns a new FlashData.
func NewFlash() *FlashData {
	return &FlashData{
		Data: make(map[string]string),
	}
}

// Set message to flash.
func (fd *FlashData) Set(key string, msg string, args ...interface{}) {
	if len(args) == 0 {
		fd.Data[key] = msg
	} else {
		fd.Data[key] = fmt.Sprintf(msg, args...)
	}
}

// WriteFlash store the flash data into the current session.
func (c *Context) WriteFlash(fd *FlashData) error {
	jsonValue, err := json.Marshal(fd.Data)
	if err != nil {
		return err
	}
	return c.Session.Set("framework_flash", jsonValue)
}

// GetFlash read the flash data from session.
func (c *Context) GetFlash() (map[string]string, error) {
	var values map[string]string
	if err := json.Unmarshal(c.Session.Get("framework_flash").([]byte), &values); err != nil {
		return nil, err
	}

	return values, nil
}
