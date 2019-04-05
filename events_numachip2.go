package main

import (
   "unsafe"
   "golang.org/x/sys/unix"
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
   statsLen       = 0x590
   venDevId       = 0x07001b47
   venDev         = 0x0000 / 4
   statCountTotal = 0x3050 / 4
   statCtrl       = 0x3058 / 4
   statCounters   = 0x3100 / 4
   wrapLimit      = 0xffffffffffff // 48 bits

   // stats counters
   statElapsed   = 0x000 / 8
)

var (
   numachip2Events = []Event{
      {0x000, false, "n2Cyc", "cycles"},
      {0x008, false, "n2CycRmpeHalf", "cycles at least half of the available RMPE contexts were in use"},
      {0x010, false, "n2CycRmpe0freeS", "cycles the RMPE did have free contexts for SIU accesses"},
      {0x030, false, "n2CycRmpe0freeP", "cycles the RMPE did have free contexts for PIU accesses"},
      {0x050, false, "n2ReqPiuRmpe", "requests from PIU to RMPE"},
      {0x058, false, "n2ValidCycReqPiuRmpe", "valid cycles acked for requests from PIU to RMPE"},
      {0x060, false, "n2WaitCycReqPiuRmpe", "wait cycles for requests from from PIU to RMPE"},
      {0x068, false, "n2ResPiuRmpe", "responses from PIU to RMPE"},
      {0x070, false, "n2ValidCycResPiuRmpe", "valid cycles acked for responses from PIU to RMPE"},
      {0x078, false, "n2WaitCycResPiuRmpe", "wait cycles for responses from from PIU to RMPE"},
      {0x080, false, "n2ReqSiuRmpe", "requests from SIU to RMPE"},
      {0x088, false, "n2CycReqSiuRmpe", "valid cycles acked for requests from SIU to RMPE"},
      {0x090, false, "n2WaitCycReqSiuRmpe", "wait cycles for requests from from SIU to RMPE"},
      {0x098, false, "n2RespSiuRmpe", "responses from SIU to RMPE"},
      {0x0A0, false, "n2ValidCycAckRespSiuRmpe", "valid cycles acked for responses from SIU to RMPE"},
      {0x0A8, false, "n2WaitCycRespSiuRmpe", "wait cycles for responses from from SIU to RMPE"},
      {0x0B0, false, "n2CycHalfLmpeUsed", "cycles at least half of the available LMPE contexts were in use"},
      {0x0B8, false, "n2CycLmpeFreePiu", "cycles the LMPE did have free contexts for SIU accesses"},
      {0x0D8, false, "n2CycLmpeFreeSiu", "cycles the LMPE did have free contexts for PIU accesses"},
      {0x0F8, false, "n2ReqPiuLmpe", "requests from PIU to LMPE"},
      {0x100, false, "n2WaitcycReqPiuLmpe", "wait cycles for requests from from PIU to LMPE"},
      {0x108, false, "n2RespPiuLmpe", "responses from PIU to LMPE"},
      {0x110, false, "n2ValidCycResPiuLmpe", "valid cycles acked for responses from PIU to LMPE"},
      {0x118, false, "n2WaitCycRespPiuLmpe", "wait cycles for responses from from PIU to LMPE"},
      {0x120, false, "n2ReqSiuLmpe", "requests from SIU to LMPE"},
      {0x128, false, "n2ValidCycAckReqSiuLmpe", "valid cycles acked for requests from SIU to LMPE"},
      {0x130, false, "n2WaitCycReqSiuLmpe", "wait cycles for requests from from SIU to LMPE"},
      {0x138, false, "n2RespSiuLmpe", "responses from SIU to LMPE"},
      {0x140, false, "n2ValidCycAckRespSiuLmpe", "valid cycles acked for responses from SIU to LMPE"},
      {0x148, false, "n2WaitCycRespSiuLmpe", "wait cycles for responses from from SIU to LMPE"},
      {0x210, false, "n2VicBlkXRecv", "VicBlk and VicBlkClean commands received on Hypertransport"},
      {0x218, false, "n2RdBlkXRecv", "RdBlk and RdBlkS commands received on Hypertransport"},
      {0x220, false, "n2RdBlkModRecv", "RdBlkMod commands received on Hypertransport"},
      {0x228, false, "n2ChangeToDirtyRecv", "ChangeToDirty commands received on Hypertransport"},
      {0x230, false, "n2RdSizedRecv", "RdSized commands received on Hypertransport"},
      {0x238, false, "n2WrSizedRecv", "WrSized commands received on Hypertransport"},
      {0x240, false, "n2DirPrbRecv", "directed Probe commands received on Hypertransport"},
      {0x248, false, "n2BcastPrbRecv", "broadcast Probe commands received on Hypertransport"},
      {0x250, false, "n2BcastCmdRecv", "Broadcast commands received on Hypertransport"},
      {0x258, false, "n2RdRespCmdRecv", "RdResponse commands received on Hypertransport"},
      {0x260, false, "n2PrbRespCmdRecv", "ProbeResponse commands received on Hypertransport"},
      {0x268, false, "n2CachelinesRecv", "data packets with full cachelines of data received on Hypertransport"},
      {0x270, false, "n2PartCachelinesRecv", "data packets with less than a full cache line received on Hypertransport"},
      {0x278, false, "n2VicBlkXSent", "VicBlk and VicBlkClean commands sent on Hypertransport"},
      {0x280, false, "n2RdBlkXSent", "RdBlk and RdBlkS commands sent on Hypertransport"},
      {0x288, false, "n2RdBlkModSent", "RdBlkMod commands sent on Hypertransport"},
      {0x290, false, "n2ChangeToDirtySent", "ChangeToDirty commands sent on Hypertransport"},
      {0x298, false, "n2RdSizedSent", "RdSized commands sent on Hypertransport"},
      {0x2A0, false, "n2WrSizedSent", "WrSized commands sent on Hypertransport"},
      {0x2A8, false, "n2BcastProbeCmdSent", "broadcast Probe commands sent on Hypertransport"},
      {0x2B0, false, "n2BcastCmdSent", "Broadcast commands sent on Hypertransport"},
      {0x2B8, false, "n2RdRespSent", "RdResponse commands sent on Hypertransport"},
      {0x2C0, false, "n2ProbeRespSent", "ProbeResponse commands sent on Hypertransport"},
      {0x2C8, false, "n2CachelinesSent", "data packets with full cachelines of data sent on Hypertransport"},
      {0x2D0, false, "n2LessCachelinesSent", "data packets with less than a full cache line sent on Hypertransport"},
      {0x2D8, false, "n2CacheReadHitRmpe", "nCache read hits on RMPE"},
      {0x2E0, false, "n2CacheStoreHitRmpe", "nCache store hits on RMPE"},
      {0x2E8, false, "n2CacheStoreMissRmpe", "nCache store misses on RMPE"},
      {0x2F0, false, "n2CacheRolloutRmpe", "nCache roll outs on RMPE"},
      {0x2F8, false, "n2CacaheInvalidatesRmpe", "nCache invalidates on RMPE"},
      {0x378, false, "n2CycOneFreeHreqPiu", "cycles with at least one free Hreq context in PIU"},
      {0x380, false, "n2CycOneFreePprb", "cycles with at least one free Pprb context in PIU"},
      {0x388, false, "n2CycOneFreeHprb", "cycles with at least one free Hprb context in PIU"},
      {0x390, false, "n2cycOneFreePtre", "cycles with at least one free Preq context in PIU"},
      {0x398, false, "n2CacheTag0Accesses", "accesses to Ctag cache 0"},
      {0x3A0, false, "n2CacheTag0WriteHit", "write hit accesses to Ctag cache 0"},
      {0x3A8, false, "n2CacheTag0ReadHit", "read hit accesses to Ctag cache 0"},
      {0x3B0, false, "n2CacheTag0WriteWriteback", "write accesses with writebacks to Ctag cache 0"},
      {0x3B8, false, "n2CacheTag0ReadWriteback", "read accesses with writebacks to Ctag cache 0"},
      {0x3C0, false, "n2CacheTag0WriteMiss", "write miss accesses to Ctag cache 0"},
      {0x3C8, false, "n2CacheTag0ReadMiss", "read miss accesses to Ctag cache 0"},
      {0x3D0, false, "n2CacheTag1Accesses", "accesses to Ctag cache 1"},
      {0x3D8, false, "n2CacheTag1WriteHit", "write hit accesses to Ctag cache 1"},
      {0x3E0, false, "n2CacheTag1ReadHit", "read hit accesses to Ctag cache 1"},
      {0x3E8, false, "n2CacheTag1WriteWriteback", "write accesses with writebacks to Ctag cache 1"},
      {0x3F0, false, "n2CacheTag1ReadWriteback", "read accesses with writebacks to Ctag cache 1"},
      {0x3F8, false, "n2CacheTag1WriteMiss", "write miss accesses to Ctag cache 1"},
      {0x400, false, "n2CacheTag1ReadMiss", "read miss accesses to Ctag cache 1"},
      {0x408, false, "n2CacheTag2Access", "accesses to Ctag cache 2"},
      {0x410, false, "n2CacheTag2WriteHit", "write hit accesses to Ctag cache 2"},
      {0x418, false, "n2CacheTag2ReadHit", "read hit accesses to Ctag cache 2"},
      {0x420, false, "n2CacheTag2WriteWritebacks", "write accesses with writebacks to Ctag cache 2"},
      {0x428, false, "n2CacheTag2WriteWritebacks", "read accesses with writebacks to Ctag cache 2"},
      {0x430, false, "n2CacheTag2WriteMiss", "write miss accesses to Ctag cache 2"},
      {0x438, false, "n2CacheTag2WriteMiss", "read miss accesses to Ctag cache 2"},
      {0x440, false, "n2CacheTag3Accesses", "accesses to Ctag cache 3"},
      {0x448, false, "n2CacheTag3WriteHit", "write hit accesses to Ctag cache 3"},
      {0x450, false, "n2CacheTag3ReadHit", "read hit accesses to Ctag cache 3"},
      {0x458, false, "n2CacheTag3WriteWriteback", "write accesses with writebacks to Ctag cache 3"},
      {0x460, false, "n2CacheTag3ReadWritebacks", "read accesses with writebacks to Ctag cache 3"},
      {0x468, false, "n2CacheTag3WriteMiss", "write miss accesses to Ctag cache 3"},
      {0x470, false, "n2CacheTag3ReadMiss", "read miss accesses to Ctag cache 3"},
      {0x478, false, "n2CacheTag4Access", "accesses to Ctag cache 4"},
      {0x480, false, "n2CacheTag4WriteHit", "write hit accesses to Ctag cache 4"},
      {0x488, false, "n2CacheTag4ReadHit", "read hit accesses to Ctag cache 4"},
      {0x490, false, "n2CacheTag4WriteWritebacks", "write accesses with writebacks to Ctag cache 4"},
      {0x498, false, "n2CacheTag4ReadWritebacks", "read accesses with writebacks to Ctag cache 4"},
      {0x4A0, false, "n2CacheTag4WriteMiss", "write miss accesses to Ctag cache 4"},
      {0x4A8, false, "n2CacheTag4ReadMiss", "read miss accesses to Ctag cache 4"},
      {0x4B0, false, "n2MainTag0Access", "accesses to Mtag cache 0"},
      {0x4B8, false, "n2MainTag0WriteHit", "write hit accesses to Mtag cache 0"},
      {0x4C0, false, "n2MainTag0ReadHit", "read hit accesses to Mtag cache 0"},
      {0x4C8, false, "n2MainTag0WriteWriteback", "write accesses with writebacks to Mtag cache 0"},
      {0x4D0, false, "n2MainTag0ReadWriteback", "read accesses with writebacks to Mtag cache 0"},
      {0x4D8, false, "n2MainTag0WriteMiss", "write miss accesses to Mtag cache 0"},
      {0x4E0, false, "n2MainTag0ReadMiss", "read miss accesses to Mtag cache 0"},
      {0x4E8, false, "n2MainTag1Access", "accesses to Mtag cache 1"},
      {0x4F0, false, "n2MainTag1WriteHit", "write hit accesses to Mtag cache 1"},
      {0x4F8, false, "n2MainTag1ReadHit", "read hit accesses to Mtag cache 1"},
      {0x500, false, "n2MainTag1WriteWriteback", "write accesses with writebacks to Mtag cache 1"},
      {0x508, false, "n2MainTag1ReadWriteback", "read accesses with writebacks to Mtag cache 1"},
      {0x510, false, "n2MainTag1WriteMiss", "write miss accesses to Mtag cache 1"},
      {0x518, false, "n2MainTag1ReadMiss", "read miss accesses to Mtag cache 1"},
      {0x520, false, "n2MainTag2Access", "accesses to Mtag cache 2"},
      {0x528, false, "n2MainTag2WriteHit", "write hit accesses to Mtag cache 2"},
      {0x530, false, "n2MainTag2ReadHit", "read hit accesses to Mtag cache 2"},
      {0x538, false, "n2MainTag2WriteWriteback", "write accesses with writebacks to Mtag cache 2"},
      {0x540, false, "n2MainTag2ReadWriteback", "read accesses with writebacks to Mtag cache 2"},
      {0x548, false, "n2MainTag2WriteMiss", "write miss accesses to Mtag cache 2"},
      {0x550, false, "n2MainTag2ReadMiss", "read miss accesses to Mtag cache 2"},
      {0x558, false, "n2MainTag3Access", "accesses to Mtag cache 3"},
      {0x560, false, "n2MainTag3WriteHit", "write hit accesses to Mtag cache 3"},
      {0x568, false, "n2MainTag3ReadHit", "read hit accesses to Mtag cache 3"},
      {0x570, false, "n2MainTag3WriteAccess", "write accesses with writebacks to Mtag cache 3"},
      {0x578, false, "n2MainTag3ReadAccess", "read accesses with writebacks to Mtag cache 3"},
      {0x580, false, "n2MainTag3WriteMiss", "write miss accesses to Mtag cache 3"},
      {0x588, false, "n2MainTag3ReadMiss", "read miss accesses to Mtag cache 3"},
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

   val := d.stats[statElapsed]
   var interval uint64 // in units of 5ns

   // if wrapped, add remainder
   if val < d.lastElapsed {
      interval = val + (wrapLimit - val)
   } else {
      interval = val - d.lastElapsed
   }

   d.lastElapsed = val

   for i, offset := range d.events {
      val = d.stats[numachip2Events[offset].index/8]
      var delta uint64

      // if wrapped, add remainder
      if val < d.last[i] {
         delta = val + (wrapLimit - val)
      } else {
         delta = val - d.last[i]
      }

      samples[i] = delta * 200000000 / interval // clockcycles @ 200MHz
      d.last[i] = val
   }

   d.regs[statCtrl] = 1 | (1 << 2) // reenable counting
   return samples
}
