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
   "syscall"
)

type Event struct {
   index    int16 // -1 means unindexed
   mnemonic string
   desc     string
   enabled  bool
}

type Sensor interface {
   // human-readable name of hardware
   Name() string
   // checks if hardware is present
   Present() bool
   // maximum sample value for percentages
   Rate() uint
   // number of hardware elements detected
   Sources() uint
   // supported events
   Events() []Event
   // gets names of enabled events
   Enable(discrete bool)
   // gets names of enabled events
   Headings(mnemonic bool) []string
   // returns samples
   Sample() []int64
   // used to prevent hardware access races
   Lock()
   Unlock()
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

func validateNonblock(err error) {
   errno, ok := err.(syscall.Errno)

   if ok && errno != syscall.EAGAIN {
      _, file, line, _ := runtime.Caller(1)
      _, leaf := path.Split(file)
      fmt.Printf("Failed with '%v' at %v:%v\n", err, leaf, line)
      os.Exit(1)
   }
}
