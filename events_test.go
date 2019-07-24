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
   "fmt"
   "testing"
)

func TestMain(m *testing.M) {
   fmt.Println("TestMain")
   dev := &Numaconnect2{}

   if dev.Present() {
      events := dev.Events()
      events[1].enabled = true
      events[3].enabled = true

      dev.Enable(true)

      for i := 0; i < 3; i++ {
         _ = dev.Sample()
      }
   } else {
      fmt.Println("Numachip2 not detected")
   }
}
