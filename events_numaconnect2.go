package main

import (
   "fmt"
   "unsafe"
   "golang.org/x/sys/unix"
)

type Numachip2 struct {
   regs        *[mapLen / 4]uint32
   stats       *[statsLen / 8]uint64
   last        []uint64
   lastElapsed uint64
}

type Numaconnect2 struct {
   events   []Event
   cards    []Numachip2
   discrete bool
   nEnabled int
}

const (
   mapBase        = 0xf0000000
   mapLen         = 0x4000
   statsLen       = 0x590
   venDevId       = 0x07001b47
   venDev         = 0x0000 / 4
   info           = 0x1090 / 4
   statCountTotal = 0x3050 / 4
   statCtrl       = 0x3058 / 4
   statCounters   = 0x3100 / 4
   wrapLimit      = 0xffffffffffff // 48 bits

   // stats counters
   statElapsed    = 0x000 / 8
)

func NewNumaconnect2() *Numaconnect2 {
   return &Numaconnect2{
      events: []Event{
         {0x000, "n2Cyc", "cycles", false},
         {0x008, "n2CycRmpeHalf", "cycles at least half of the available RMPE contexts were in use", false},
         {0x010, "n2CycRmpe0freeS", "cycles the RMPE did have free contexts for SIU accesses", false},
         {0x030, "n2CycRmpe0freeP", "cycles the RMPE did have free contexts for PIU accesses", false},
         {0x050, "n2ReqPiuRmpe", "requests from PIU to RMPE", false},
         {0x058, "n2ValidCycReqPiuRmpe", "valid cycles acked for requests from PIU to RMPE", false},
         {0x060, "n2WaitCycReqPiuRmpe", "wait cycles for requests from from PIU to RMPE", false},
         {0x068, "n2ResPiuRmpe", "responses from PIU to RMPE", false},
         {0x070, "n2ValidCycResPiuRmpe", "valid cycles acked for responses from PIU to RMPE", false},
         {0x078, "n2WaitCycResPiuRmpe", "wait cycles for responses from from PIU to RMPE", false},
         {0x080, "n2ReqSiuRmpe", "requests from SIU to RMPE", false},
         {0x088, "n2CycReqSiuRmpe", "valid cycles acked for requests from SIU to RMPE", false},
         {0x090, "n2WaitCycReqSiuRmpe", "wait cycles for requests from from SIU to RMPE", false},
         {0x098, "n2RespSiuRmpe", "responses from SIU to RMPE", false},
         {0x0A0, "n2ValidCycAckRespSiuRmpe", "valid cycles acked for responses from SIU to RMPE", false},
         {0x0A8, "n2WaitCycRespSiuRmpe", "wait cycles for responses from from SIU to RMPE", false},
         {0x0B0, "n2CycHalfLmpeUsed", "cycles at least half of the available LMPE contexts were in use", false},
         {0x0B8, "n2CycLmpeFreePiu", "cycles the LMPE did have free contexts for SIU accesses", false},
         {0x0D8, "n2CycLmpeFreeSiu", "cycles the LMPE did have free contexts for PIU accesses", false},
         {0x0F8, "n2ReqPiuLmpe", "requests from PIU to LMPE", false},
         {0x100, "n2WaitcycReqPiuLmpe", "wait cycles for requests from from PIU to LMPE", false},
         {0x108, "n2RespPiuLmpe", "responses from PIU to LMPE", false},
         {0x110, "n2ValidCycResPiuLmpe", "valid cycles acked for responses from PIU to LMPE", false},
         {0x118, "n2WaitCycRespPiuLmpe", "wait cycles for responses from from PIU to LMPE", false},
         {0x120, "n2ReqSiuLmpe", "requests from SIU to LMPE", false},
         {0x128, "n2ValidCycAckReqSiuLmpe", "valid cycles acked for requests from SIU to LMPE", false},
         {0x130, "n2WaitCycReqSiuLmpe", "wait cycles for requests from from SIU to LMPE", false},
         {0x138, "n2RespSiuLmpe", "responses from SIU to LMPE", false},
         {0x140, "n2ValidCycAckRespSiuLmpe", "valid cycles acked for responses from SIU to LMPE", false},
         {0x148, "n2WaitCycRespSiuLmpe", "wait cycles for responses from from SIU to LMPE", false},
         {0x210, "n2VicBlkXRecv", "VicBlk and VicBlkClean commands received on Hypertransport", false},
         {0x218, "n2RdBlkXRecv", "RdBlk and RdBlkS commands received on Hypertransport", false},
         {0x220, "n2RdBlkModRecv", "RdBlkMod commands received on Hypertransport", false},
         {0x228, "n2ChangeToDirtyRecv", "ChangeToDirty commands received on Hypertransport", false},
         {0x230, "n2RdSizedRecv", "RdSized commands received on Hypertransport", false},
         {0x238, "n2WrSizedRecv", "WrSized commands received on Hypertransport", false},
         {0x240, "n2DirPrbRecv", "directed Probe commands received on Hypertransport", false},
         {0x248, "n2BcastPrbRecv", "broadcast Probe commands received on Hypertransport", false},
         {0x250, "n2BcastCmdRecv", "Broadcast commands received on Hypertransport", false},
         {0x258, "n2RdRespCmdRecv", "RdResponse commands received on Hypertransport", false},
         {0x260, "n2PrbRespCmdRecv", "ProbeResponse commands received on Hypertransport", false},
         {0x268, "n2CachelinesRecv", "data packets with full cachelines of data received on Hypertransport", false},
         {0x270, "n2PartCachelinesRecv", "data packets with less than a full cache line received on Hypertransport", false},
         {0x278, "n2VicBlkXSent", "VicBlk and VicBlkClean commands sent on Hypertransport", false},
         {0x280, "n2RdBlkXSent", "RdBlk and RdBlkS commands sent on Hypertransport", false},
         {0x288, "n2RdBlkModSent", "RdBlkMod commands sent on Hypertransport", false},
         {0x290, "n2ChangeToDirtySent", "ChangeToDirty commands sent on Hypertransport", false},
         {0x298, "n2RdSizedSent", "RdSized commands sent on Hypertransport", false},
         {0x2A0, "n2WrSizedSent", "WrSized commands sent on Hypertransport", false},
         {0x2A8, "n2BcastProbeCmdSent", "broadcast Probe commands sent on Hypertransport", false},
         {0x2B0, "n2BcastCmdSent", "Broadcast commands sent on Hypertransport", false},
         {0x2B8, "n2RdRespSent", "RdResponse commands sent on Hypertransport", false},
         {0x2C0, "n2ProbeRespSent", "ProbeResponse commands sent on Hypertransport", false},
         {0x2C8, "n2CachelinesSent", "data packets with full cachelines of data sent on Hypertransport", false},
         {0x2D0, "n2LessCachelinesSent", "data packets with less than a full cache line sent on Hypertransport", false},
         {0x2D8, "n2CacheReadHitRmpe", "nCache read hits on RMPE", false},
         {0x2E0, "n2CacheStoreHitRmpe", "nCache store hits on RMPE", false},
         {0x2E8, "n2CacheStoreMissRmpe", "nCache store misses on RMPE", false},
         {0x2F0, "n2CacheRolloutRmpe", "nCache roll outs on RMPE", false},
         {0x2F8, "n2CacaheInvalidatesRmpe", "nCache invalidates on RMPE", false},
         {0x378, "n2CycOneFreeHreqPiu", "cycles with at least one free Hreq context in PIU", false},
         {0x380, "n2CycOneFreePprb", "cycles with at least one free Pprb context in PIU", false},
         {0x388, "n2CycOneFreeHprb", "cycles with at least one free Hprb context in PIU", false},
         {0x390, "n2cycOneFreePtre", "cycles with at least one free Preq context in PIU", false},
         {0x398, "n2CacheTag0Accesses", "accesses to Ctag cache 0", false},
         {0x3A0, "n2CacheTag0WriteHit", "write hit accesses to Ctag cache 0", false},
         {0x3A8, "n2CacheTag0ReadHit", "read hit accesses to Ctag cache 0", false},
         {0x3B0, "n2CacheTag0WriteWriteback", "write accesses with writebacks to Ctag cache 0", false},
         {0x3B8, "n2CacheTag0ReadWriteback", "read accesses with writebacks to Ctag cache 0", false},
         {0x3C0, "n2CacheTag0WriteMiss", "write miss accesses to Ctag cache 0", false},
         {0x3C8, "n2CacheTag0ReadMiss", "read miss accesses to Ctag cache 0", false},
         {0x3D0, "n2CacheTag1Accesses", "accesses to Ctag cache 1", false},
         {0x3D8, "n2CacheTag1WriteHit", "write hit accesses to Ctag cache 1", false},
         {0x3E0, "n2CacheTag1ReadHit", "read hit accesses to Ctag cache 1", false},
         {0x3E8, "n2CacheTag1WriteWriteback", "write accesses with writebacks to Ctag cache 1", false},
         {0x3F0, "n2CacheTag1ReadWriteback", "read accesses with writebacks to Ctag cache 1", false},
         {0x3F8, "n2CacheTag1WriteMiss", "write miss accesses to Ctag cache 1", false},
         {0x400, "n2CacheTag1ReadMiss", "read miss accesses to Ctag cache 1", false},
         {0x408, "n2CacheTag2Access", "accesses to Ctag cache 2", false},
         {0x410, "n2CacheTag2WriteHit", "write hit accesses to Ctag cache 2", false},
         {0x418, "n2CacheTag2ReadHit", "read hit accesses to Ctag cache 2", false},
         {0x420, "n2CacheTag2WriteWritebacks", "write accesses with writebacks to Ctag cache 2", false},
         {0x428, "n2CacheTag2ReadWritebacks", "read accesses with writebacks to Ctag cache 2", false},
         {0x430, "n2CacheTag2WriteMiss", "write miss accesses to Ctag cache 2", false},
         {0x438, "n2CacheTag2ReadMiss", "read miss accesses to Ctag cache 2", false},
         {0x440, "n2CacheTag3Accesses", "accesses to Ctag cache 3", false},
         {0x448, "n2CacheTag3WriteHit", "write hit accesses to Ctag cache 3", false},
         {0x450, "n2CacheTag3ReadHit", "read hit accesses to Ctag cache 3", false},
         {0x458, "n2CacheTag3WriteWriteback", "write accesses with writebacks to Ctag cache 3", false},
         {0x460, "n2CacheTag3ReadWritebacks", "read accesses with writebacks to Ctag cache 3", false},
         {0x468, "n2CacheTag3WriteMiss", "write miss accesses to Ctag cache 3", false},
         {0x470, "n2CacheTag3ReadMiss", "read miss accesses to Ctag cache 3", false},
         {0x478, "n2CacheTag4Access", "accesses to Ctag cache 4", false},
         {0x480, "n2CacheTag4WriteHit", "write hit accesses to Ctag cache 4", false},
         {0x488, "n2CacheTag4ReadHit", "read hit accesses to Ctag cache 4", false},
         {0x490, "n2CacheTag4WriteWritebacks", "write accesses with writebacks to Ctag cache 4", false},
         {0x498, "n2CacheTag4ReadWritebacks", "read accesses with writebacks to Ctag cache 4", false},
         {0x4A0, "n2CacheTag4WriteMiss", "write miss accesses to Ctag cache 4", false},
         {0x4A8, "n2CacheTag4ReadMiss", "read miss accesses to Ctag cache 4", false},
         {0x4B0, "n2MainTag0Access", "accesses to Mtag cache 0", false},
         {0x4B8, "n2MainTag0WriteHit", "write hit accesses to Mtag cache 0", false},
         {0x4C0, "n2MainTag0ReadHit", "read hit accesses to Mtag cache 0", false},
         {0x4C8, "n2MainTag0WriteWriteback", "write accesses with writebacks to Mtag cache 0", false},
         {0x4D0, "n2MainTag0ReadWriteback", "read accesses with writebacks to Mtag cache 0", false},
         {0x4D8, "n2MainTag0WriteMiss", "write miss accesses to Mtag cache 0", false},
         {0x4E0, "n2MainTag0ReadMiss", "read miss accesses to Mtag cache 0", false},
         {0x4E8, "n2MainTag1Access", "accesses to Mtag cache 1", false},
         {0x4F0, "n2MainTag1WriteHit", "write hit accesses to Mtag cache 1", false},
         {0x4F8, "n2MainTag1ReadHit", "read hit accesses to Mtag cache 1", false},
         {0x500, "n2MainTag1WriteWriteback", "write accesses with writebacks to Mtag cache 1", false},
         {0x508, "n2MainTag1ReadWriteback", "read accesses with writebacks to Mtag cache 1", false},
         {0x510, "n2MainTag1WriteMiss", "write miss accesses to Mtag cache 1", false},
         {0x518, "n2MainTag1ReadMiss", "read miss accesses to Mtag cache 1", false},
         {0x520, "n2MainTag2Access", "accesses to Mtag cache 2", false},
         {0x528, "n2MainTag2WriteHit", "write hit accesses to Mtag cache 2", false},
         {0x530, "n2MainTag2ReadHit", "read hit accesses to Mtag cache 2", false},
         {0x538, "n2MainTag2WriteWriteback", "write accesses with writebacks to Mtag cache 2", false},
         {0x540, "n2MainTag2ReadWriteback", "read accesses with writebacks to Mtag cache 2", false},
         {0x548, "n2MainTag2WriteMiss", "write miss accesses to Mtag cache 2", false},
         {0x550, "n2MainTag2ReadMiss", "read miss accesses to Mtag cache 2", false},
         {0x558, "n2MainTag3Access", "accesses to Mtag cache 3", false},
         {0x560, "n2MainTag3WriteHit", "write hit accesses to Mtag cache 3", false},
         {0x568, "n2MainTag3ReadHit", "read hit accesses to Mtag cache 3", false},
         {0x570, "n2MainTag3WriteAccess", "write accesses with writebacks to Mtag cache 3", false},
         {0x578, "n2MainTag3ReadAccess", "read accesses with writebacks to Mtag cache 3", false},
         {0x580, "n2MainTag3WriteMiss", "write miss accesses to Mtag cache 3", false},
         {0x588, "n2MainTag3ReadMiss", "read miss accesses to Mtag cache 3", false},
      },
   }
}

func (d *Numaconnect2) Present() bool {
   fd, err := unix.Open("/dev/mem", unix.O_RDWR, 0)
   validate(err)

   data, err := unix.Mmap(fd, mapBase, mapLen, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED|unix.MAP_FILE)
   validate(err)
   defer unix.Munmap(data)

   regs := (*[mapLen/4]uint32)(unsafe.Pointer(&data[0]))
   if regs[venDev] != venDevId {
      return false
   }

   master := (regs[info+5] >> 4) & 0xfff
   hts := (regs[info+6] >> 12) & 7

   for pos := master; pos != 0xfff; {
      base := 0x3f0000000000 | (int64(pos) << 28) | ((23+int64(hts)) << 15)

      data, err := unix.Mmap(fd, base, mapLen, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED|unix.MAP_FILE)
      validate(err)

      regs := (*[mapLen/4]uint32)(unsafe.Pointer(&data[0]))
      if regs[venDev] != venDevId {
         fmt.Printf("vendev %08x\n", regs[venDev])
         panic("mismatching vendev")
      }

      stats := (*[statsLen / 8]uint64)(unsafe.Pointer(&regs[statCounters]))
      d.cards = append(d.cards, Numachip2{regs: regs, stats: stats})

      pos = regs[info+6] & 0xfff
   }

   return true
}

func (d *Numaconnect2) Name() string {
   return "Numascale NumaConnect2"
}

func (d *Numaconnect2) Enable(discrete bool) {
   d.discrete = discrete
   d.nEnabled = 0

   for _, event := range d.events {
      if event.enabled {
         d.nEnabled++
      }
   }

   for i := range d.cards {
      d.cards[i].regs[statCtrl] = 0            // reset block
      d.cards[i].regs[statCtrl] = 1 | (1 << 2) // enable counting
      d.cards[i].last = make([]uint64, d.nEnabled)
   }
}

func (d *Numaconnect2) Headings() []string {
   var headings []string

   for _, event := range d.events {
      if !event.enabled {
         continue
      }

      if d.discrete {
         for i := 0; i < len(d.cards); i++ {
            heading := fmt.Sprintf("%s:%d", event.mnemonic, i)
            headings = append(headings, heading)
         }
      } else {
         headings = append(headings, event.mnemonic)
      }
   }

   return headings
}

func (d *Numaconnect2) Sample() []int64 {
   var samples []int64

   if d.discrete {
      samples = make([]int64, d.nEnabled * len(d.cards))
   } else {
      samples = make([]int64, d.nEnabled)
   }

   for n := range d.cards {
      d.cards[n].regs[statCtrl] = 1 // disable counting

      val := d.cards[n].stats[statElapsed]
      var interval uint64 // in units of 5ns

      // if wrapped, add remainder
      if val < d.cards[n].lastElapsed {
         interval = val + (wrapLimit - val)
      } else {
         interval = val - d.cards[n].lastElapsed
      }

      d.cards[n].lastElapsed = val
      i := 0

      for _, event := range d.events {
         if !event.enabled {
            continue
         }

         val = d.cards[n].stats[event.index/8]
         var delta uint64

         // if wrapped, add remainder
         if val < d.cards[n].last[i] {
            delta = val + (wrapLimit - val)
         } else {
            delta = val - d.cards[n].last[i]
         }

         sample := delta * 200000000 / interval // clockcycles @ 200MHz

         if d.discrete {
            samples[n*len(d.events)+i] = int64(sample)
         } else {
            // sum totals for average
            samples[i] += int64(sample)
            d.cards[n].last[i] = val
         }

         i++
      }

      d.cards[n].regs[statCtrl] = 1 | (1 << 2) // reenable counting
   }

   if !d.discrete {
      // divide through for average
      for i := range samples {
         samples[i] /= int64(len(d.cards))
      }
   }

   return samples
}

func (d *Numaconnect2) Events() []Event {
   return d.events
}