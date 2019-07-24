/*  Copyright (C) 2019 Daniel J Blueman
    This file is part of Numascope.

    Numascope is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Numascope is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Numascope.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
   "os"
   "fmt"
   "path"
   "runtime"
)

type Event struct {
   index    int16 // -1 means unindexed
   mnemonic string
   desc     string
   enabled  bool
}

type Sensor interface {
   // checks if sensor is present
   Present() bool
   Name() string
   Sources() uint
   // scans through and activates enabled events, and if discrete
   Enable(discrete bool)
   Lock()
   Unlock()
   // returns slice of events
   Events() []Event
   // returns headings of enabled events, accounting for discrete or not
   Headings() []string
   // returns samples
   Sample() []int64
}


// Checks if an error occurred
func validate(err error) {
   if err != nil {
      _, file, line, _ := runtime.Caller(1)
      _, leaf := path.Split(file)
      fmt.Printf("Failed with '%v' at %v:%v\n", err, leaf, line)
      os.Exit(1)
   }
}
