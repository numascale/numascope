package main

import (
   "fmt"
   "sync"
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
   mutex    sync.Mutex
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
//         {0x000/8, "n2Cyc", "cycles", false},
         {0x008/8, "n2CycRmpeHalf", "cycles at least half of the available RMPE contexts were in use", false},
         {0x010/8, "n2CycRmpe0freeS", "cycles the RMPE had free contexts for SIU accesses", false},
         {0x030/8, "n2CycRmpe0freeP", "cycles the RMPE had free contexts for PIU accesses", false},
         {0x050/8, "n2ReqPiuRmpe", "requests from PIU to RMPE", false},
         {0x058/8, "n2ValidCycReqPiuRmpe", "valid cycles acked for requests from PIU to RMPE", false},
         {0x060/8, "n2WaitCycReqPiuRmpe", "wait cycles for requests from from PIU to RMPE", false},
         {0x068/8, "n2ResPiuRmpe", "responses from PIU to RMPE", false},
         {0x070/8, "n2ValidCycResPiuRmpe", "valid cycles acked for responses from PIU to RMPE", false},
         {0x078/8, "n2WaitCycResPiuRmpe", "wait cycles for responses from from PIU to RMPE", false},
         {0x080/8, "n2ReqSiuRmpe", "requests from SIU to RMPE", false},
         {0x088/8, "n2CycReqSiuRmpe", "valid cycles acked for requests from SIU to RMPE", false},
         {0x090/8, "n2WaitCycReqSiuRmpe", "wait cycles for requests from from SIU to RMPE", false},
         {0x098/8, "n2RespSiuRmpe", "responses from SIU to RMPE", false},
         {0x0A0/8, "n2ValidCycAckRespSiuRmpe", "valid cycles acked for responses from SIU to RMPE", false},
         {0x0A8/8, "n2WaitCycRespSiuRmpe", "wait cycles for responses from from SIU to RMPE", false},
         {0x0B0/8, "n2CycHalfLmpeUsed", "cycles at least half of the available LMPE contexts were in use", false},
         {0x0B8/8, "n2CycLmpeFreePiu", "cycles the LMPE had free contexts for SIU accesses", false},
         {0x0D8/8, "n2CycLmpeFreeSiu", "cycles the LMPE had free contexts for PIU accesses", false},
         {0x0F8/8, "n2ReqPiuLmpe", "requests from PIU to LMPE", false},
         {0x100/8, "n2WaitcycReqPiuLmpe", "wait cycles for requests from from PIU to LMPE", false},
         {0x108/8, "n2RespPiuLmpe", "responses from PIU to LMPE", false},
         {0x110/8, "n2ValidCycResPiuLmpe", "valid cycles acked for responses from PIU to LMPE", false},
         {0x118/8, "n2WaitCycRespPiuLmpe", "wait cycles for responses from from PIU to LMPE", false},
         {0x120/8, "n2ReqSiuLmpe", "requests from SIU to LMPE", false},
         {0x128/8, "n2ValidCycAckReqSiuLmpe", "valid cycles acked for requests from SIU to LMPE", false},
         {0x130/8, "n2WaitCycReqSiuLmpe", "wait cycles for requests from from SIU to LMPE", false},
         {0x138/8, "n2RespSiuLmpe", "responses from SIU to LMPE", false},
         {0x140/8, "n2ValidCycAckRespSiuLmpe", "valid cycles acked for responses from SIU to LMPE", false},
         {0x148/8, "n2WaitCycRespSiuLmpe", "wait cycles for responses from from SIU to LMPE", false},
         {0x210/8, "n2VicBlkXRecv", "VicBlk and VicBlkClean commands received", false},
         {0x218/8, "n2RdBlkXRecv", "RdBlk and RdBlkS commands received", false},
         {0x220/8, "n2RdBlkModRecv", "RdBlkMod commands received", false},
         {0x228/8, "n2ChangeToDirtyRecv", "ChangeToDirty commands received", false},
         {0x230/8, "n2RdSizedRecv", "RdSized commands received", false},
         {0x238/8, "n2WrSizedRecv", "WrSized commands received", false},
         {0x240/8, "n2DirPrbRecv", "directed Probe commands received", false},
         {0x248/8, "n2BcastPrbRecv", "broadcast Probe commands received", false},
         {0x250/8, "n2BcastCmdRecv", "Broadcast commands received", false},
         {0x258/8, "n2RdRespCmdRecv", "RdResponse commands received", false},
         {0x260/8, "n2PrbRespCmdRecv", "ProbeResponse commands received", false},
         {0x268/8, "n2CachelinesRecv", "data packets with full cachelines of data received", false},
         {0x270/8, "n2PartCachelinesRecv", "data packets with less than a full cache line received", false},
         {0x278/8, "n2VicBlkXSent", "VicBlk and VicBlkClean commands sent", false},
         {0x280/8, "n2RdBlkXSent", "RdBlk and RdBlkS commands sent", false},
         {0x288/8, "n2RdBlkModSent", "RdBlkMod commands sent", false},
         {0x290/8, "n2ChangeToDirtySent", "ChangeToDirty commands sent", false},
         {0x298/8, "n2RdSizedSent", "RdSized commands sent", false},
         {0x2A0/8, "n2WrSizedSent", "WrSized commands sent", false},
         {0x2A8/8, "n2BcastProbeCmdSent", "broadcast Probe commands sent", false},
         {0x2B0/8, "n2BcastCmdSent", "broadcast commands sent", false},
         {0x2B8/8, "n2RdRespSent", "RdResponse commands sent", false},
         {0x2C0/8, "n2ProbeRespSent", "ProbeResponse commands sent", false},
         {0x2C8/8, "n2CachelinesSent", "data packets with full cachelines of data sent", false},
         {0x2D0/8, "n2LessCachelinesSent", "data packets with less than a full cache line sent", false},
         {0x2D8/8, "n2CacheReadHitRmpe", "nCache read hits on RMPE", false},
         {0x2E0/8, "n2CacheStoreHitRmpe", "nCache store hits on RMPE", false},
         {0x2E8/8, "n2CacheStoreMissRmpe", "nCache store misses on RMPE", false},
         {0x2F0/8, "n2CacheRolloutRmpe", "nCache roll outs on RMPE", false},
         {0x2F8/8, "n2CacaheInvalidatesRmpe", "nCache invalidates on RMPE", false},
         {0x378/8, "n2CycOneFreeHreqPiu", "cycles with at least one free Hreq context in PIU", false},
         {0x380/8, "n2CycOneFreePprb", "cycles with at least one free Pprb context in PIU", false},
         {0x388/8, "n2CycOneFreeHprb", "cycles with at least one free Hprb context in PIU", false},
         {0x390/8, "n2cycOneFreePtre", "cycles with at least one free Preq context in PIU", false},
         {0x398/8, "n2CacheTag0Accesses", "accesses to Ctag cache 0", false},
         {0x3A0/8, "n2CacheTag0WriteHit", "write hit accesses to Ctag cache 0", false},
         {0x3A8/8, "n2CacheTag0ReadHit", "read hit accesses to Ctag cache 0", false},
         {0x3B0/8, "n2CacheTag0WriteWriteback", "write accesses with writebacks to Ctag cache 0", false},
         {0x3B8/8, "n2CacheTag0ReadWriteback", "read accesses with writebacks to Ctag cache 0", false},
         {0x3C0/8, "n2CacheTag0WriteMiss", "write miss accesses to Ctag cache 0", false},
         {0x3C8/8, "n2CacheTag0ReadMiss", "read miss accesses to Ctag cache 0", false},
         {0x3D0/8, "n2CacheTag1Accesses", "accesses to Ctag cache 1", false},
         {0x3D8/8, "n2CacheTag1WriteHit", "write hit accesses to Ctag cache 1", false},
         {0x3E0/8, "n2CacheTag1ReadHit", "read hit accesses to Ctag cache 1", false},
         {0x3E8/8, "n2CacheTag1WriteWriteback", "write accesses with writebacks to Ctag cache 1", false},
         {0x3F0/8, "n2CacheTag1ReadWriteback", "read accesses with writebacks to Ctag cache 1", false},
         {0x3F8/8, "n2CacheTag1WriteMiss", "write miss accesses to Ctag cache 1", false},
         {0x400/8, "n2CacheTag1ReadMiss", "read miss accesses to Ctag cache 1", false},
         {0x408/8, "n2CacheTag2Access", "accesses to Ctag cache 2", false},
         {0x410/8, "n2CacheTag2WriteHit", "write hit accesses to Ctag cache 2", false},
         {0x418/8, "n2CacheTag2ReadHit", "read hit accesses to Ctag cache 2", false},
         {0x420/8, "n2CacheTag2WriteWritebacks", "write accesses with writebacks to Ctag cache 2", false},
         {0x428/8, "n2CacheTag2ReadWritebacks", "read accesses with writebacks to Ctag cache 2", false},
         {0x430/8, "n2CacheTag2WriteMiss", "write miss accesses to Ctag cache 2", false},
         {0x438/8, "n2CacheTag2ReadMiss", "read miss accesses to Ctag cache 2", false},
         {0x440/8, "n2CacheTag3Accesses", "accesses to Ctag cache 3", false},
         {0x448/8, "n2CacheTag3WriteHit", "write hit accesses to Ctag cache 3", false},
         {0x450/8, "n2CacheTag3ReadHit", "read hit accesses to Ctag cache 3", false},
         {0x458/8, "n2CacheTag3WriteWriteback", "write accesses with writebacks to Ctag cache 3", false},
         {0x460/8, "n2CacheTag3ReadWritebacks", "read accesses with writebacks to Ctag cache 3", false},
         {0x468/8, "n2CacheTag3WriteMiss", "write miss accesses to Ctag cache 3", false},
         {0x470/8, "n2CacheTag3ReadMiss", "read miss accesses to Ctag cache 3", false},
         {0x4B0/8, "n2MainTag0Access", "accesses to Mtag cache 0", false},
         {0x4B8/8, "n2MainTag0WriteHit", "write hit accesses to Mtag cache 0", false},
         {0x4C0/8, "n2MainTag0ReadHit", "read hit accesses to Mtag cache 0", false},
         {0x4C8/8, "n2MainTag0WriteWriteback", "write accesses with writebacks to Mtag cache 0", false},
         {0x4D0/8, "n2MainTag0ReadWriteback", "read accesses with writebacks to Mtag cache 0", false},
         {0x4D8/8, "n2MainTag0WriteMiss", "write miss accesses to Mtag cache 0", false},
         {0x4E0/8, "n2MainTag0ReadMiss", "read miss accesses to Mtag cache 0", false},
         {0x4E8/8, "n2MainTag1Access", "accesses to Mtag cache 1", false},
         {0x4F0/8, "n2MainTag1WriteHit", "write hit accesses to Mtag cache 1", false},
         {0x4F8/8, "n2MainTag1ReadHit", "read hit accesses to Mtag cache 1", false},
         {0x500/8, "n2MainTag1WriteWriteback", "write accesses with writebacks to Mtag cache 1", false},
         {0x508/8, "n2MainTag1ReadWriteback", "read accesses with writebacks to Mtag cache 1", false},
         {0x510/8, "n2MainTag1WriteMiss", "write miss accesses to Mtag cache 1", false},
         {0x518/8, "n2MainTag1ReadMiss", "read miss accesses to Mtag cache 1", false},
         {0x520/8, "n2MainTag2Access", "accesses to Mtag cache 2", false},
         {0x528/8, "n2MainTag2WriteHit", "write hit accesses to Mtag cache 2", false},
         {0x530/8, "n2MainTag2ReadHit", "read hit accesses to Mtag cache 2", false},
         {0x538/8, "n2MainTag2WriteWriteback", "write accesses with writebacks to Mtag cache 2", false},
         {0x540/8, "n2MainTag2ReadWriteback", "read accesses with writebacks to Mtag cache 2", false},
         {0x548/8, "n2MainTag2WriteMiss", "write miss accesses to Mtag cache 2", false},
         {0x550/8, "n2MainTag2ReadMiss", "read miss accesses to Mtag cache 2", false},
         {0x558/8, "n2MainTag3Access", "accesses to Mtag cache 3", false},
         {0x560/8, "n2MainTag3WriteHit", "write hit accesses to Mtag cache 3", false},
         {0x568/8, "n2MainTag3ReadHit", "read hit accesses to Mtag cache 3", false},
         {0x570/8, "n2MainTag3WriteAccess", "write accesses with writebacks to Mtag cache 3", false},
         {0x578/8, "n2MainTag3ReadAccess", "read accesses with writebacks to Mtag cache 3", false},
         {0x580/8, "n2MainTag3WriteMiss", "write miss accesses to Mtag cache 3", false},
         {0x588/8, "n2MainTag3ReadMiss", "read miss accesses to Mtag cache 3", false},
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

func (d *Numaconnect2) Lock() {
   d.mutex.Lock()
}

func (d *Numaconnect2) Unlock() {
   d.mutex.Unlock()
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

   d.Lock()
   defer d.Unlock()

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

         val = d.cards[n].stats[event.index]
         var delta uint64

         // if wrapped, add remainder
         if val < d.cards[n].last[i] {
            delta = val + (wrapLimit - val)
         } else {
            delta = val - d.cards[n].last[i]
         }

         sample := delta * 200000000 / interval // clockcycles @ 200MHz

         if d.discrete {
            samples[n*d.nEnabled+i] = int64(sample)
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
