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
   events      []uint16 // index into event list
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

var (
   numachip2Events = []Event{
      {0x000, false, "", "Total number of 5ns clock cycles since statistics was enabled"},
      {0x008, false, "", "Number of cycles at least half of the available RMPE contexts were in use"},
      {0x010, false, "", "Number of cycles the RMPE[0] did have free contexts for SIU accesses"},
      {0x018, false, "", "Number of cycles the RMPE[1] did have free contexts for SIU accesses"},
      {0x020, false, "", "Number of cycles the RMPE[2] did have free contexts for SIU accesses"},
      {0x028, false, "", "Number of cycles the RMPE[3] did have free contexts for SIU accesses"},
      {0x030, false, "", "Number of cycles the RMPE[0] did have free contexts for PIU accesses"},
      {0x038, false, "", "Number of cycles the RMPE[1] did have free contexts for PIU accesses"},
      {0x040, false, "", "Number of cycles the RMPE[2] did have free contexts for PIU accesses"},
      {0x048, false, "", "Number of cycles the RMPE[3] did have free contexts for PIU accesses"},
      {0x050, false, "", "Number of requests from PIU to RMPE"},
      {0x058, false, "", "Number of valid cycles acked for requestes from PIU to RMPE"},
      {0x060, false, "", "Number of wait cycles for requests from from PIU to RMPE"},
      {0x068, false, "", "Number of responses from PIU to RMPE"},
      {0x070, false, "", "Number of valid cycles acked for responses from PIU to RMPE"},
      {0x078, false, "", "Number of wait cycles for responses from from PIU to RMPE"},
      {0x080, false, "", "Number of requests from SIU to RMPE"},
      {0x088, false, "", "Number of valid cycles acked for requestes from SIU to RMPE"},
      {0x090, false, "", "Number of wait cycles for requests from from SIU to RMPE"},
      {0x098, false, "", "Number of responses from SIU to RMPE"},
      {0x0A0, false, "", "Number of valid cycles acked for responses from SIU to RMPE"},
      {0x0A8, false, "", "Number of wait cycles for responses from from SIU to RMPE"},
      {0x0B0, false, "", "Number of cycles at least half of the available LMPE contexts were in use"},
      {0x0B8, false, "", "Number of cycles the LMPE[0] did have free contexts for SIU accesses"},
      {0x0C0, false, "", "Number of cycles the LMPE[1] did have free contexts for SIU accesses"},
      {0x0C8, false, "", "Number of cycles the LMPE[2] did have free contexts for SIU accesses"},
      {0x0D0, false, "", "Number of cycles the LMPE[3] did have free contexts for SIU accesses"},
      {0x0D8, false, "", "Number of cycles the LMPE[0] did have free contexts for PIU accesses"},
      {0x0E0, false, "", "Number of cycles the LMPE[1] did have free contexts for PIU accesses"},
      {0x0E8, false, "", "Number of cycles the LMPE[2] did have free contexts for PIU accesses"},
      {0x0F0, false, "", "Number of cycles the LMPE[3] did have free contexts for PIU accesses"},
      {0x0F8, false, "", "Number of requests from PIU to LMPE"},
      {0x100, false, "", "Number of wait cycles for requests from from PIU to LMPE"},
      {0x108, false, "", "Number of responses from PIU to LMPE"},
      {0x110, false, "", "Number of valid cycles acked for responses from PIU to LMPE"},
      {0x118, false, "", "Number of wait cycles for responses from from PIU to LMPE"},
      {0x120, false, "", "Number of requests from SIU to LMPE"},
      {0x128, false, "", "Number of valid cycles acked for requestes from SIU to LMPE"},
      {0x130, false, "", "Number of wait cycles for requests from from SIU to LMPE"},
      {0x138, false, "", "Number of responses from SIU to LMPE"},
      {0x140, false, "", "Number of valid cycles acked for responses from SIU to LMPE"},
      {0x148, false, "", "Number of wait cycles for responses from from SIU to LMPE"},
      {0x210, false, "", "Number of VicBlk and VicBlkClean commands received on Hypertransport"},
      {0x218, false, "", "Number of RdBlk and RdBlkS commands received on Hypertransport"},
      {0x220, false, "", "Number of RdBlkMod commands received on Hypertransport"},
      {0x228, false, "", "Number of ChangeToDirty commands received on Hypertransport"},
      {0x230, false, "", "Number of RdSized commands received on Hypertransport"},
      {0x238, false, "", "Number of WrSized commands received on Hypertransport"},
      {0x240, false, "", "Number of directed Probe commands received on Hypertransport"},
      {0x248, false, "", "Number of broadcast Probe commands received on Hypertransport"},
      {0x250, false, "", "Number of Broadcast commands received on Hypertransport"},
      {0x258, false, "", "Number of RdResponse commands received on Hypertransport"},
      {0x260, false, "", "Number of ProbeResponse commands received on Hypertransport"},
      {0x268, false, "", "Number of data packets with full cachelines of data received on Hypertransport"},
      {0x270, false, "", "Number of data packets with less than a full cache line received on Hypertransport"},
      {0x278, false, "", "Number of VicBlk and VicBlkClean commands sent on Hypertransport"},
      {0x280, false, "", "Number of RdBlk and RdBlkS commands sent on Hypertransport"},
      {0x288, false, "", "Number of RdBlkMod commands sent on Hypertransport"},
      {0x290, false, "", "Number of ChangeToDirty commands sent on Hypertransport"},
      {0x298, false, "", "Number of RdSized commands sent on Hypertransport"},
      {0x2A0, false, "", "Number of WrSized commands sent on Hypertransport"},
      {0x2A8, false, "", "Number of broadcast Probe commands sent on Hypertransport"},
      {0x2B0, false, "", "Number of Broadcast commands sent on Hypertransport"},
      {0x2B8, false, "", "Number of RdResponse commands sent on Hypertransport"},
      {0x2C0, false, "", "Number of ProbeResponse commands sent on Hypertransport"},
      {0x2C8, false, "", "Number of data packets with full cachelines of data sent on Hypertransport"},
      {0x2D0, false, "", "Number of data packets with less than a full cache line sent on Hypertransport"},
      {0x2D8, false, "", "Number of nache read hits on RMPE[0]"},
      {0x2E0, false, "", "Number of ncahe store hits on RMPE[0]"},
      {0x2E8, false, "", "Number of nache store misses on RMPE[0]"},
      {0x2F0, false, "", "Number of nache roll outs on RMPE[0]"},
      {0x2F8, false, "", "Number of ncahe invalides on RMPE[0]"},
      {0x300, false, "", "Number of nache read hits on RMPE[1]"},
      {0x308, false, "", "Number of ncahe store hits on RMPE[1]"},
      {0x310, false, "", "Number of nache store misses on RMPE[1]"},
      {0x318, false, "", "Number of nache roll outs on RMPE[1]"},
      {0x320, false, "", "Number of nache invalides on RMPE[1]"},
      {0x328, false, "", "Number of nache read hits on RMPE[2]"},
      {0x330, false, "", "Number of ncahe store hits on RMPE[2]"},
      {0x338, false, "", "Number of nache store misses on RMPE[2]"},
      {0x340, false, "", "Number of nache roll outs on RMPE[2]"},
      {0x348, false, "", "Number of nache invalides on RMPE[2]"},
      {0x350, false, "", "Number of nache read hits on RMPE[3]"},
      {0x358, false, "", "Number of ncahe store hits on RMPE[3]"},
      {0x360, false, "", "Number of nache store misses on RMPE[3]"},
      {0x368, false, "", "Number of nache roll outs on RMPE[3]"},
      {0x370, false, "", "Number of nache invalides on RMPE[3]"},
      {0x378, false, "", "Number if clock cycles with at least one free Hreq context in PIU"},
      {0x380, false, "", "Number if clock cycles with at least one free Pprb context in PIU"},
      {0x388, false, "", "Number if clock cycles with at least one free Hprb context in PIU"},
      {0x390, false, "", "Number if clock cycles with at least one free Preq context in PIU"},
      {0x398, false, "", "Number of accesses to Ctag cache 0"},
      {0x3A0, false, "", "Number of write hit accesses to Ctag cache 0"},
      {0x3A8, false, "", "Number of read hit accesses to Ctag cache 0"},
      {0x3B0, false, "", "Number of write accesses with write backs to Ctag cache 0"},
      {0x3B8, false, "", "Number of read accesses with write backs to Ctag cache 0"},
      {0x3C0, false, "", "Number of write miss accesses to Ctag cache 0"},
      {0x3C8, false, "", "Number of read miss accesses to Ctag cache 0"},
      {0x3D0, false, "", "Number of accesses to Ctag cache 1"},
      {0x3D8, false, "", "Number of write hit accesses to Ctag cache 1"},
      {0x3E0, false, "", "Number of read hit accesses to Ctag cache 1"},
      {0x3E8, false, "", "Number of write accesses with write backs to Ctag cache 1"},
      {0x3F0, false, "", "Number of read accesses with write backs to Ctag cache 1"},
      {0x3F8, false, "", "Number of write miss accesses to Ctag cache 1"},
      {0x400, false, "", "Number of read miss accesses to Ctag cache 1"},
      {0x408, false, "", "Number of accesses to Ctag cache 2"},
      {0x410, false, "", "Number of write hit accesses to Ctag cache 2"},
      {0x418, false, "", "Number of read hit accesses to Ctag cache 2"},
      {0x420, false, "", "Number of write accesses with write backs to Ctag cache 2"},
      {0x428, false, "", "Number of read accesses with write backs to Ctag cache 2"},
      {0x430, false, "", "Number of write miss accesses to Ctag cache 2"},
      {0x438, false, "", "Number of read miss accesses to Ctag cache 2"},
      {0x440, false, "", "Number of accesses to Ctag cache 3"},
      {0x448, false, "", "Number of write hit accesses to Ctag cache 3"},
      {0x450, false, "", "Number of read hit accesses to Ctag cache 3"},
      {0x458, false, "", "Number of write accesses with write backs to Ctag cache 3"},
      {0x460, false, "", "Number of read accesses with write backs to Ctag cache 3"},
      {0x468, false, "", "Number of write miss accesses to Ctag cache 3"},
      {0x470, false, "", "Number of read miss accesses to Ctag cache 3"},
      {0x478, false, "", "Number of accesses to Ctag cache 4"},
      {0x480, false, "", "Number of write hit accesses to Ctag cache 4"},
      {0x488, false, "", "Number of read hit accesses to Ctag cache 4"},
      {0x490, false, "", "Number of write accesses with write backs to Ctag cache 4"},
      {0x498, false, "", "Number of read accesses with write backs to Ctag cache 4"},
      {0x4A0, false, "", "Number of write miss accesses to Ctag cache 4"},
      {0x4A8, false, "", "Number of read miss accesses to Ctag cache 4"},
      {0x4B0, false, "", "Number of accesses to Mtag cache 0"},
      {0x4B8, false, "", "Number of write hit accesses to Mtag cache 0"},
      {0x4C0, false, "", "Number of read hit accesses to Mtag cache 0"},
      {0x4C8, false, "", "Number of write accesses with write backs to Mtag cache 0"},
      {0x4D0, false, "", "Number of read accesses with write backs to Mtag cache 0"},
      {0x4D8, false, "", "Number of write miss accesses to Mtag cache 0"},
      {0x4E0, false, "", "Number of read miss accesses to Mtag cache 0"},
      {0x4E8, false, "", "Number of accesses to Mtag cache 1"},
      {0x4F0, false, "", "Number of write hit accesses to Mtag cache 1"},
      {0x4F8, false, "", "Number of read hit accesses to Mtag cache 1"},
      {0x500, false, "", "Number of write accesses with write backs to Mtag cache 1"},
      {0x508, false, "", "Number of read accesses with write backs to Mtag cache 1"},
      {0x510, false, "", "Number of write miss accesses to Mtag cache 1"},
      {0x518, false, "", "Number of read miss accesses to Mtag cache 1"},
      {0x520, false, "", "Number of accesses to Mtag cache 2"},
      {0x528, false, "", "Number of write hit accesses to Mtag cache 2"},
      {0x530, false, "", "Number of read hit accesses to Mtag cache 2"},
      {0x538, false, "", "Number of write accesses with write backs to Mtag cache 2"},
      {0x540, false, "", "Number of read accesses with write backs to Mtag cache 2"},
      {0x548, false, "", "Number of write miss accesses to Mtag cache 2"},
      {0x550, false, "", "Number of read miss accesses to Mtag cache 2"},
      {0x558, false, "", "Number of accesses to Mtag cache 3"},
      {0x560, false, "", "Number of write hit accesses to Mtag cache 3"},
      {0x568, false, "", "Number of read hit accesses to Mtag cache 3"},
      {0x570, false, "", "Number of write accesses with write backs to Mtag cache 3"},
      {0x578, false, "", "Number of read accesses with write backs to Mtag cache 3"},
      {0x580, false, "", "Number of write miss accesses to Mtag cache 3"},
      {0x588, false, "", "Number of read miss accesses to Mtag cache 3"},
   }
)

func (d *Numachip2) probe() *[]Event {
   fd, err := unix.Open("/dev/mem", unix.O_RDWR, 0)
   validate(err)
   defer unix.Close(fd)

   data, err := unix.Mmap(fd, mapBase, mapLen, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED|unix.MAP_FILE)
   validate(err)

   d.regs = (*[mapLen / 4]uint32)(unsafe.Pointer(&data[0]))
   d.stats = (*[statsLen / 8]uint64)(unsafe.Pointer(&d.regs[statCounters]))

   if d.regs[venDev] == venDevId {
      return &numachip2Events
   }

   return &[]Event{}
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
