package main

import (
   "fmt"
   "golang.org/x/sys/unix"
   "unsafe"
)

type EventsNumachip2 interface {
   probe() bool
   enable([]uint16)
   sample() []uint64
}

type Numachip2 struct {
   regs        *[mapLen / 4]uint32
   stats       *[statsLen / 8]uint64
   events      []uint16
   last        []uint64
   lastElapsed uint64
}

const (
   mapBase        = 0xf0000000
   mapLen         = 0x4000
   statsLen       = 0x1000
   venDevId       = 0x07001b47
   venDev         = 0x0000 / 4
   statCountTotal = 0x3050 / 4
   statCtrl       = 0x3058 / 4
   statCounters   = 0x3100 / 4

   // stats counters
   statElapsed   = 0x000 / 8
   statRdBlkXRec = 0x218 / 8
)

func (d *Numachip2) probe() bool {
   fd, err := unix.Open("/dev/mem", unix.O_RDWR, 0)
   validate(err)
   defer unix.Close(fd)

   data, err := unix.Mmap(fd, mapBase, mapLen, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED|unix.MAP_FILE)
   validate(err)

   d.regs = (*[mapLen / 4]uint32)(unsafe.Pointer(&data[0]))
   d.stats = (*[statsLen / 8]uint64)(unsafe.Pointer(&d.regs[statCounters]))

   return d.regs[venDev] == venDevId
}

func (d *Numachip2) enable(events []uint16) {
   d.events = events
   d.regs[statCtrl] = 0            // reset block
   d.regs[statCtrl] = 1 | (1 << 2) // enable counting

   d.last = make([]uint64, len(events))
}

func (d *Numachip2) sample() []uint64 {
   samples := make([]uint64, len(d.events))
   d.regs[statCtrl] = 1 // disable counting

   dElapsed := d.stats[statElapsed]
   var interval uint64 // in units of 5ns

   // if wrapped, add remainder
   if dElapsed < d.last[statElapsed] {
      interval = dElapsed + (0xffffffffffffffff - dElapsed)
   } else {
      interval = dElapsed - d.lastElapsed
   }

   d.lastElapsed = dElapsed

   // FIXME handle wrapping
   for i, offset := range d.events {
      val := d.stats[offset]
      fmt.Printf("i %v, interval %v, offset %v, val %v, last %v\n", i, interval, offset, val, d.last[i])
      samples[i] = (val - d.last[i]) * 200000000 / interval
      d.last[i] = val
   }

   d.regs[statCtrl] = 1 | (1 << 2) // reenable counting
   return samples
}
