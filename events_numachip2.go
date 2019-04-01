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

type Event struct {
   index uint16
   desc string
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

func (d *Numachip2) supported() []Event {
   return []Event{
      {0x000, "Total number of 5ns clock cycles since statistics was enabled"},
      {0x008, "Number of cycles at least half of the available RMPE contexts were in use"},
      {0x010, "Number of cycles the RMPE[0] did have free contexts for SIU accesses"},
      {0x018, "Number of cycles the RMPE[1] did have free contexts for SIU accesses"},
      {0x020, "Number of cycles the RMPE[2] did have free contexts for SIU accesses"},
      {0x028, "Number of cycles the RMPE[3] did have free contexts for SIU accesses"},
      {0x030, "Number of cycles the RMPE[0] did have free contexts for PIU accesses"},
      {0x038, "Number of cycles the RMPE[1] did have free contexts for PIU accesses"},
      {0x040, "Number of cycles the RMPE[2] did have free contexts for PIU accesses"},
      {0x048, "Number of cycles the RMPE[3] did have free contexts for PIU accesses"},
      {0x050, "Number of requests from PIU to RMPE"},
      {0x058, "Number of valid cycles acked for requestes from PIU to RMPE"},
      {0x060, "Number of wait cycles for requests from from PIU to RMPE"},
      {0x068, "Number of responses from PIU to RMPE"},
      {0x070, "Number of valid cycles acked for responses from PIU to RMPE"},
      {0x078, "Number of wait cycles for responses from from PIU to RMPE"},
      {0x080, "Number of requests from SIU to RMPE"},
      {0x088, "Number of valid cycles acked for requestes from SIU to RMPE"},
      {0x090, "Number of wait cycles for requests from from SIU to RMPE"},
      {0x098, "Number of responses from SIU to RMPE"},
      {0x0A0, "Number of valid cycles acked for responses from SIU to RMPE"},
      {0x0A8, "Number of wait cycles for responses from from SIU to RMPE"},
      {0x0B0, "Number of cycles at least half of the available LMPE contexts were in use"},
      {0x0B8, "Number of cycles the LMPE[0] did have free contexts for SIU accesses"},
      {0x0C0, "Number of cycles the LMPE[1] did have free contexts for SIU accesses"},
      {0x0C8, "Number of cycles the LMPE[2] did have free contexts for SIU accesses"},
      {0x0D0, "Number of cycles the LMPE[3] did have free contexts for SIU accesses"},
      {0x0D8, "Number of cycles the LMPE[0] did have free contexts for PIU accesses"},
      {0x0E0, "Number of cycles the LMPE[1] did have free contexts for PIU accesses"},
      {0x0E8, "Number of cycles the LMPE[2] did have free contexts for PIU accesses"},
      {0x0F0, "Number of cycles the LMPE[3] did have free contexts for PIU accesses"},
      {0x0F8, "Number of requests from PIU to LMPE"},
      {0x100, "Number of wait cycles for requests from from PIU to LMPE"},
      {0x108, "Number of responses from PIU to LMPE"},
      {0x110, "Number of valid cycles acked for responses from PIU to LMPE"},
      {0x118, "Number of wait cycles for responses from from PIU to LMPE"},
      {0x120, "Number of requests from SIU to LMPE"},
      {0x128, "Number of valid cycles acked for requestes from SIU to LMPE"},
      {0x130, "Number of wait cycles for requests from from SIU to LMPE"},
      {0x138, "Number of responses from SIU to LMPE"},
      {0x140, "Number of valid cycles acked for responses from SIU to LMPE"},
      {0x148, "Number of wait cycles for responses from from SIU to LMPE"},
      {0x210, "Number of VicBlk and VicBlkClean commands received on Hypertransport"},
      {0x218, "Number of RdBlk and RdBlkS commands received on Hypertransport"},
      {0x220, "Number of RdBlkMod commands received on Hypertransport"},
      {0x228, "Number of ChangeToDirty commands received on Hypertransport"},
      {0x230, "Number of RdSized commands received on Hypertransport"},
      {0x238, "Number of WrSized commands received on Hypertransport"},
      {0x240, "Number of directed Probe commands received on Hypertransport"},
      {0x248, "Number of broadcast Probe commands received on Hypertransport"},
      {0x250, "Number of Broadcast commands received on Hypertransport"},
      {0x258, "Number of RdResponse commands received on Hypertransport"},
      {0x260, "Number of ProbeResponse commands received on Hypertransport"},
      {0x268, "Number of data packets with full cachelines of data received on Hypertransport"},
      {0x270, "Number of data packets with less than a full cache line received on Hypertransport"},
      {0x278, "Number of VicBlk and VicBlkClean commands sent on Hypertransport"},
      {0x280, "Number of RdBlk and RdBlkS commands sent on Hypertransport"},
      {0x288, "Number of RdBlkMod commands sent on Hypertransport"},
      {0x290, "Number of ChangeToDirty commands sent on Hypertransport"},
      {0x298, "Number of RdSized commands sent on Hypertransport"},
      {0x2A0, "Number of WrSized commands sent on Hypertransport"},
      {0x2A8, "Number of broadcast Probe commands sent on Hypertransport"},
      {0x2B0, "Number of Broadcast commands sent on Hypertransport"},
      {0x2B8, "Number of RdResponse commands sent on Hypertransport"},
      {0x2C0, "Number of ProbeResponse commands sent on Hypertransport"},
      {0x2C8, "Number of data packets with full cachelines of data sent on Hypertransport"},
      {0x2D0, "Number of data packets with less than a full cache line sent on Hypertransport"},
      {0x2D8, "Number of nache read hits on RMPE[0]"},
      {0x2E0, "Number of ncahe store hits on RMPE[0]"},
      {0x2E8, "Number of nache store misses on RMPE[0]"},
      {0x2F0, "Number of nache roll outs on RMPE[0]"},
      {0x2F8, "Number of ncahe invalides on RMPE[0]"},
      {0x300, "Number of nache read hits on RMPE[1]"},
      {0x308, "Number of ncahe store hits on RMPE[1]"},
      {0x310, "Number of nache store misses on RMPE[1]"},
      {0x318, "Number of nache roll outs on RMPE[1]"},
      {0x320, "Number of nache invalides on RMPE[1]"},
      {0x328, "Number of nache read hits on RMPE[2]"},
      {0x330, "Number of ncahe store hits on RMPE[2]"},
      {0x338, "Number of nache store misses on RMPE[2]"},
      {0x340, "Number of nache roll outs on RMPE[2]"},
      {0x348, "Number of nache invalides on RMPE[2]"},
      {0x350, "Number of nache read hits on RMPE[3]"},
      {0x358, "Number of ncahe store hits on RMPE[3]"},
      {0x360, "Number of nache store misses on RMPE[3]"},
      {0x368, "Number of nache roll outs on RMPE[3]"},
      {0x370, "Number of nache invalides on RMPE[3]"},
      {0x378, "Number if clock cycles with at least one free Hreq context in PIU"},
      {0x380, "Number if clock cycles with at least one free Pprb context in PIU"},
      {0x388, "Number if clock cycles with at least one free Hprb context in PIU"},
      {0x390, "Number if clock cycles with at least one free Preq context in PIU"},
      {0x398, "Number of accesses to Ctag cache 0"},
      {0x3A0, "Number of write hit accesses to Ctag cache 0"},
      {0x3A8, "Number of read hit accesses to Ctag cache 0"},
      {0x3B0, "Number of write accesses with write backs to Ctag cache 0"},
      {0x3B8, "Number of read accesses with write backs to Ctag cache 0"},
      {0x3C0, "Number of write miss accesses to Ctag cache 0"},
      {0x3C8, "Number of read miss accesses to Ctag cache 0"},
      {0x3D0, "Number of accesses to Ctag cache 1"},
      {0x3D8, "Number of write hit accesses to Ctag cache 1"},
      {0x3E0, "Number of read hit accesses to Ctag cache 1"},
      {0x3E8, "Number of write accesses with write backs to Ctag cache 1"},
      {0x3F0, "Number of read accesses with write backs to Ctag cache 1"},
      {0x3F8, "Number of write miss accesses to Ctag cache 1"},
      {0x400, "Number of read miss accesses to Ctag cache 1"},
      {0x408, "Number of accesses to Ctag cache 2"},
      {0x410, "Number of write hit accesses to Ctag cache 2"},
      {0x418, "Number of read hit accesses to Ctag cache 2"},
      {0x420, "Number of write accesses with write backs to Ctag cache 2"},
      {0x428, "Number of read accesses with write backs to Ctag cache 2"},
      {0x430, "Number of write miss accesses to Ctag cache 2"},
      {0x438, "Number of read miss accesses to Ctag cache 2"},
      {0x440, "Number of accesses to Ctag cache 3"},
      {0x448, "Number of write hit accesses to Ctag cache 3"},
      {0x450, "Number of read hit accesses to Ctag cache 3"},
      {0x458, "Number of write accesses with write backs to Ctag cache 3"},
      {0x460, "Number of read accesses with write backs to Ctag cache 3"},
      {0x468, "Number of write miss accesses to Ctag cache 3"},
      {0x470, "Number of read miss accesses to Ctag cache 3"},
      {0x478, "Number of accesses to Ctag cache 4"},
      {0x480, "Number of write hit accesses to Ctag cache 4"},
      {0x488, "Number of read hit accesses to Ctag cache 4"},
      {0x490, "Number of write accesses with write backs to Ctag cache 4"},
      {0x498, "Number of read accesses with write backs to Ctag cache 4"},
      {0x4A0, "Number of write miss accesses to Ctag cache 4"},
      {0x4A8, "Number of read miss accesses to Ctag cache 4"},
      {0x4B0, "Number of accesses to Mtag cache 0"},
      {0x4B8, "Number of write hit accesses to Mtag cache 0"},
      {0x4C0, "Number of read hit accesses to Mtag cache 0"},
      {0x4C8, "Number of write accesses with write backs to Mtag cache 0"},
      {0x4D0, "Number of read accesses with write backs to Mtag cache 0"},
      {0x4D8, "Number of write miss accesses to Mtag cache 0"},
      {0x4E0, "Number of read miss accesses to Mtag cache 0"},
      {0x4E8, "Number of accesses to Mtag cache 1"},
      {0x4F0, "Number of write hit accesses to Mtag cache 1"},
      {0x4F8, "Number of read hit accesses to Mtag cache 1"},
      {0x500, "Number of write accesses with write backs to Mtag cache 1"},
      {0x508, "Number of read accesses with write backs to Mtag cache 1"},
      {0x510, "Number of write miss accesses to Mtag cache 1"},
      {0x518, "Number of read miss accesses to Mtag cache 1"},
      {0x520, "Number of accesses to Mtag cache 2"},
      {0x528, "Number of write hit accesses to Mtag cache 2"},
      {0x530, "Number of read hit accesses to Mtag cache 2"},
      {0x538, "Number of write accesses with write backs to Mtag cache 2"},
      {0x540, "Number of read accesses with write backs to Mtag cache 2"},
      {0x548, "Number of write miss accesses to Mtag cache 2"},
      {0x550, "Number of read miss accesses to Mtag cache 2"},
      {0x558, "Number of accesses to Mtag cache 3"},
      {0x560, "Number of write hit accesses to Mtag cache 3"},
      {0x568, "Number of read hit accesses to Mtag cache 3"},
      {0x570, "Number of write accesses with write backs to Mtag cache 3"},
      {0x578, "Number of read accesses with write backs to Mtag cache 3"},
      {0x580, "Number of write miss accesses to Mtag cache 3"},
      {0x588, "Number of read miss accesses to Mtag cache 3"},
   }
}

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
