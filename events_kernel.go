package main

import (
   "os"
   "strings"
   "strconv"
   "time"
)

type Kernel struct {
   events      []Event
   file        *os.File
   last        []uint64
   lastElapsed time.Time
   nEnabled    int
}

func NewKernel() *Kernel {
   return &Kernel{
      events: []Event{
         // include/linux/mmzone.h
         {-1, "nr_free_pages", "unallocated pages", false},
         {-1, "nr_zone_inactive_anon", "zone inactive anonymous pages", false},
         {-1, "nr_zone_active_anon", "zone activate anonymous pages", false},
         {-1, "nr_zone_inactive_file", "zone inactive file-backed pages", false},
         {-1, "nr_zone_active_file", "zone active file-backed pages", false},
         {-1, "nr_zone_unevictable", "zone unevictable pages", false},
         {-1, "nr_zone_write_pending", "zone write pending pages", false},
         {-1, "nr_mlock", "locked pages", false},
         {-1, "nr_page_table_pages", "page table pages", false},
         {-1, "nr_kernel_stack", "kernel stack kilobytes", false},
         {-1, "nr_bounce", "low-memory pages allocated for DMA", false},
         {-1, "nr_free_cma", "free Contig Mem Alloc pages", false},
         {-1, "numa_hit", "allocated in intended node", false},
         {-1, "numa_miss", "allocated in non-intended node", false},
         {-1, "numa_foreign", "was intended here, hit elsewhere", false},
         {-1, "numa_interleave", "interleaver preferred this zone", false},
         {-1, "numa_local", "allocation from local node", false},
         {-1, "numa_other", "allocation from non-local node", false},
         {-1, "nr_inactive_anon", "inactive anonymous pages", false},
         {-1, "nr_active_anon", "active anonymous pages", false},
         {-1, "nr_inactive_file", "inactive file-backed pages", false},
         {-1, "nr_active_file", "active file-backed pages", false},
         {-1, "nr_unevictable", "unevictable Pages", false},
         {-1, "nr_slab_reclaimable", "unused bytes allocated to slab", false},
         {-1, "nr_slab_unreclaimable", "used bytes allocated to slab", false},
         {-1, "nr_isolated_anon", "temporary anonymous isolated pages", false},
         {-1, "nr_isolated_file", "temporary file-backed isolated pages", false},
         {-1, "workingset_refault", "refaults of previously evicted pages", false},
         {-1, "workingset_activate", "refaulted pages that were immediately activated", false},
         {-1, "workingset_nodereclaim", "times a shadow node has been reclaimed", false},
         {-1, "nr_anon_pages", "non file-backed memory-mapped pages", false},
         {-1, "nr_mapped", "file-backed memory-mapped pages", false},
         {-1, "nr_file_pages", "pagecache pages", false},
         {-1, "nr_dirty", "dirty pagecache pages", false},
         {-1, "nr_writeback", "pagecache pages pending writeback", false},
         {-1, "nr_writeback_temp", "pagecache pages pending writeback using temporary buffers", false},
         {-1, "nr_shmem", "shared memory pages including tmpfs and GEM pages", false},
         {-1, "nr_shmem_hugepages", "shared memory 2MB or larger pages", false},
         {-1, "nr_shmem_pmdmapped", "shared memory pages mapped via middle directory", false},
         {-1, "nr_anon_transparent_hugepages", "non file-backed 2MB or larger pages", false},
         {-1, "nr_unstable", "uncommitted dirty network filesystem pages", false},
         {-1, "nr_vmscan_write", "pages paged out", false},
         {-1, "nr_vmscan_immediate_reclaim", "pages ready to be reclaimed", false},
         {-1, "nr_dirtied", "pages dirtied", false},
         {-1, "nr_written", "pages written to", false},
         {-1, "nr_dirty_threshold", "synchronous writeback threshold bytes", false},
         {-1, "nr_dirty_background_threshold", "asynchronous writeback threshold bytes", false},
         // include/linux/vm_event_item.h
         {-1, "pgpgin", "pageins", false},
         {-1, "pgpgout", "pageouts", false},
         {-1, "pswpin", "pages swapped in", false},
         {-1, "pswpout", "pages swapped out", false},
         {-1, "pgalloc_dma32", "page allocations, DMA32 zone", false},
         {-1, "pgalloc_normal", "page allocations per zone, normal zone", false},
         {-1, "pgalloc_movable", "page allocations per zone, movable zone", false},
         {-1, "allocstall_dma32", "Direct reclaim calls, DMA32 zone", false},
         {-1, "allocstall_normal", "Direct reclaim calls, normal zone", false},
         {-1, "allocstall_movable", "Direct reclaim calls, movable zone", false},
         {-1, "pgskip_dma32", "pages unscannable, DMA32 zone", false},
         {-1, "pgskip_normal", "pages unscannable, normal zone", false},
         {-1, "pgskip_movable", "pages unscannable, movable zone", false},
         {-1, "pgfree", "pages freed", false},
         {-1, "pgactivate", "pages marked frequently used", false},
         {-1, "pgdeactivate", "pages marked infrequently used", false},
         {-1, "pglazyfree", "pages pending asynchronous freeing", false},
         {-1, "pgfault", "pagefaults not causing IO", false},
         {-1, "pgmajfault", "pagefaults causing IO", false},
         {-1, "pglazyfreed", "pages freed asynchronously", false},
         {-1, "pgrefill", "page refills", false},
         {-1, "pgsteal_kswapd", "page steals by kswapd", false},
         {-1, "pgsteal_direct", "page steals on allocation path", false},
         {-1, "pgscan_kswapd", "pages scanned by the kswapd daemon", false},
         {-1, "pgscan_direct", "pages scanned in process context", false},
         {-1, "pgscan_direct_throttle", "pages scanned in throttled process context", false},
         {-1, "zone_reclaim_failed", "reclaim failures", false},
         {-1, "pginodesteal", "pages reclaimed via inode freeing", false},
         {-1, "slabs_scanned", "slab objects scanned", false},
         {-1, "kswapd_inodesteal", "pages reclaimed by kswapd via inode freeing", false},
         {-1, "kswapd_low_wmark_hit_quickly", "times kswapd reached low watermark quickly", false},
         {-1, "kswapd_high_wmark_hit_quickly", "times kswapd reached high watermark quickly", false},
         {-1, "pageoutrun", "kswapd calls to page reclaim", false},
         {-1, "pgrotated", "pages reused after IO", false},
         {-1, "drop_pagecache", "pagecache flushes", false},
         {-1, "drop_slab", "slab flushes", false},
         {-1, "oom_kill", "out of memory kills", false},
         {-1, "pgmigrate_success", "pages migrated", false},
         {-1, "pgmigrate_fail", "pages failed migration", false},
         {-1, "compact_migrate_scanned", "compactable pages marked for migration in process context", false},
         {-1, "compact_free_scanned", "compactable free pages scanned in process context", false},
         {-1, "compact_isolated", "compactable pages isolated in process context", false},
         {-1, "compact_stall", "page compaction stalls in process context", false},
         {-1, "compact_fail", "page compaction failures in process context", false},
         {-1, "compact_success", "compaction daemon succeeded runs", false},
         {-1, "compact_daemon_wake", "times compaction daemon was woken", false},
         {-1, "compact_daemon_migrate_scanned", "pages marked for migration by compaction daemon", false},
         {-1, "compact_daemon_free_scanned", "free pages scanned by compaction daemon", false},
         {-1, "htlb_buddy_alloc_success", "2MB or larger pages allocated", false},
         {-1, "htlb_buddy_alloc_fail", "2MB or larger pages failed allocation", false},
         {-1, "unevictable_pgs_culled", "pages which became unevictable", false},
         {-1, "unevictable_pgs_scanned", "unevictable pages scanned", false},
         {-1, "unevictable_pgs_rescued", "unevictable pages became evictable", false},
         {-1, "unevictable_pgs_mlocked", "unevictable pages locked", false},
         {-1, "unevictable_pgs_munlocked", "unevictable pages unlocked", false},
         {-1, "unevictable_pgs_cleared", "unevictable pages zeroed", false},
         {-1, "unevictable_pgs_stranded", "unevictable pages which couldn't be isolated", false},
         {-1, "thp_fault_alloc", "2MB or larger pages page-faulted", false},
         {-1, "thp_fault_fallback", "2MB or larger pages reused", false},
         {-1, "thp_collapse_alloc", "2MB or larger pages from merging", false},
         {-1, "thp_collapse_alloc_failed", "2MB or larger page merge failure", false},
         {-1, "thp_file_alloc", "2MB or larger file-backed pages allocated", false},
         {-1, "thp_file_mapped", "2MB or larger pagefaults", false},
         {-1, "thp_split_page", "2MB or larger pages split to normal pages", false},
         {-1, "thp_split_page_failed", "2MB or larger pages split failures", false},
         {-1, "thp_deferred_split_page", "2MB or larger pages with deferred split", false},
         {-1, "thp_split_pmd", "2MB or larger pages split from middle directory", false},
         {-1, "thp_split_pud", "2MB or larger pages split from upper directory", false},
         {-1, "thp_zero_page_alloc", "2MB or larger zero pages allocated", false},
         {-1, "thp_zero_page_alloc_failed", "2MB or larger zero page allocation failures", false},
         {-1, "thp_swpout", "2MB or larger pages swapped out", false},
         {-1, "thp_swpout_fallback", "2MB or larger pages swapped out as normal pages", false},
         {-1, "balloon_inflate", "pages added to page balloon", false},
         {-1, "balloon_deflate", "pages removed from page balloon", false},
         {-1, "swap_ra", "pages swapped in due to readahead", false},
         {-1, "swap_ra_hit", "pages returned from swap cache due to readahead", false},
      },
   }
}

func (d *Kernel) Present() bool {
   return true
}

func (d *Kernel) Name() string {
   return "kernel VMstat"
}

func (d *Kernel) Enable(discrete bool) {
   d.last = []uint64{}
   d.nEnabled = 0

   for _, event := range d.events {
      if event.enabled {
         d.nEnabled++
      }
   }

   d.last = make([]uint64, d.nEnabled)

   var err error
   d.file, err = os.Open("/proc/vmstat")
   validate(err)

   // update last values, discarding differences
   _ = d.Sample()
}

func (d *Kernel) Headings() []string {
   headings := []string{}

   for _, event := range d.events {
      if event.enabled {
         headings = append(headings, event.mnemonic)
      }
   }

   return headings
}

func (d *Kernel) Sample() []int64 {
   buf := make([]byte, 8192)

   current := time.Now()
   delta := uint64(current.Sub(d.lastElapsed) / time.Nanosecond)
   d.lastElapsed = current

   d.file.Seek(0, 0) // FIXME debug why SeekAt returns EOF
   n, err := d.file.Read(buf)
   validate(err)

   // parse strings into map for O(n) total cost
   m := make(map[string]uint64)
   bufs := string(buf[:n-1]) // trim trailing newline
   lines := strings.Split(bufs, "\n")

   for _, line := range lines {
      parts := strings.Split(line, " ")
      count, err := strconv.ParseUint(parts[1], 10, 64)
      validate(err)
      m[parts[0]] = count
   }

   samples := make([]int64, d.nEnabled)
   i := 0

   for _, event := range d.events {
      if !event.enabled {
         continue
      }

      val := m[event.mnemonic]
      samples[i] = (int64(val) - int64(d.last[i])) * 1000000000 / int64(delta)
      d.last[i] = val
      i++
   }

   return samples
}

func (d *Kernel) Events() []Event {
   return d.events
}
