package main

import (
   "os"
   "strings"
   "strconv"
   "time"
)

type Vmstat struct {
   enabled     []uint16 // index into event list
   file        *os.File
   last        []uint64
   lastElapsed time.Time
}

var (
   vmstatEvents = []Event{
      {-1, false, "nr_free_pages", ""},
      {-1, false, "nr_zone_inactive_anon", ""},
      {-1, false, "nr_zone_active_anon", ""},
      {-1, false, "nr_zone_inactive_file", ""},
      {-1, false, "nr_zone_active_file", ""},
      {-1, false, "nr_zone_unevictable", ""},
      {-1, false, "nr_zone_write_pending", ""},
      {-1, false, "nr_mlock", ""},
      {-1, false, "nr_page_table_pages", ""},
      {-1, false, "nr_kernel_stack", ""},
      {-1, false, "nr_bounce", ""},
      {-1, false, "nr_zspages", ""},
      {-1, false, "nr_free_cma", ""},
      {-1, false, "numa_hit", ""},
      {-1, false, "numa_miss", ""},
      {-1, false, "numa_foreign", ""},
      {-1, false, "numa_interleave", ""},
      {-1, false, "numa_local", ""},
      {-1, false, "numa_other", ""},
      {-1, false, "nr_inactive_anon", ""},
      {-1, false, "nr_active_anon", ""},
      {-1, false, "nr_inactive_file", ""},
      {-1, false, "nr_active_file", ""},
      {-1, false, "nr_unevictable", ""},
      {-1, false, "nr_slab_reclaimable", ""},
      {-1, false, "nr_slab_unreclaimable", ""},
      {-1, false, "nr_isolated_anon", ""},
      {-1, false, "nr_isolated_file", ""},
      {-1, false, "workingset_refault", ""},
      {-1, false, "workingset_activate", ""},
      {-1, false, "workingset_nodereclaim", ""},
      {-1, false, "nr_anon_pages", ""},
      {-1, false, "nr_mapped", ""},
      {-1, false, "nr_file_pages", ""},
      {-1, false, "nr_dirty", ""},
      {-1, false, "nr_writeback", ""},
      {-1, false, "nr_writeback_temp", ""},
      {-1, false, "nr_shmem", ""},
      {-1, false, "nr_shmem_hugepages", ""},
      {-1, false, "nr_shmem_pmdmapped", ""},
      {-1, false, "nr_anon_transparent_hugepages", ""},
      {-1, false, "nr_unstable", ""},
      {-1, false, "nr_vmscan_write", ""},
      {-1, false, "nr_vmscan_immediate_reclaim", ""},
      {-1, false, "nr_dirtied", ""},
      {-1, false, "nr_written", ""},
      {-1, false, "nr_dirty_threshold", ""},
      {-1, false, "nr_dirty_background_threshold", ""},
      {-1, false, "pgpgin", ""},
      {-1, false, "pgpgout", ""},
      {-1, false, "pswpin", ""},
      {-1, false, "pswpout", ""},
      {-1, false, "pgalloc_dma", ""},
      {-1, false, "pgalloc_dma32", ""},
      {-1, false, "pgalloc_normal", ""},
      {-1, false, "pgalloc_movable", ""},
      {-1, false, "allocstall_dma", ""},
      {-1, false, "allocstall_dma32", ""},
      {-1, false, "allocstall_normal", ""},
      {-1, false, "allocstall_movable", ""},
      {-1, false, "pgskip_dma", ""},
      {-1, false, "pgskip_dma32", ""},
      {-1, false, "pgskip_normal", ""},
      {-1, false, "pgskip_movable", ""},
      {-1, false, "pgfree", ""},
      {-1, false, "pgactivate", ""},
      {-1, false, "pgdeactivate", ""},
      {-1, false, "pglazyfree", ""},
      {-1, false, "pgfault", ""},
      {-1, false, "pgmajfault", ""},
      {-1, false, "pglazyfreed", ""},
      {-1, false, "pgrefill", ""},
      {-1, false, "pgsteal_kswapd", ""},
      {-1, false, "pgsteal_direct", ""},
      {-1, false, "pgscan_kswapd", ""},
      {-1, false, "pgscan_direct", ""},
      {-1, false, "pgscan_direct_throttle", ""},
      {-1, false, "zone_reclaim_failed", ""},
      {-1, false, "pginodesteal", ""},
      {-1, false, "slabs_scanned", ""},
      {-1, false, "kswapd_inodesteal", ""},
      {-1, false, "kswapd_low_wmark_hit_quickly", ""},
      {-1, false, "kswapd_high_wmark_hit_quickly", ""},
      {-1, false, "pageoutrun", ""},
      {-1, false, "pgrotated", ""},
      {-1, false, "drop_pagecache", ""},
      {-1, false, "drop_slab", ""},
      {-1, false, "oom_kill", ""},
      {-1, false, "numa_pte_updates", ""},
      {-1, false, "numa_huge_pte_updates", ""},
      {-1, false, "numa_hint_faults", ""},
      {-1, false, "numa_hint_faults_local", ""},
      {-1, false, "numa_pages_migrated", ""},
      {-1, false, "pgmigrate_success", ""},
      {-1, false, "pgmigrate_fail", ""},
      {-1, false, "compact_migrate_scanned", ""},
      {-1, false, "compact_free_scanned", ""},
      {-1, false, "compact_isolated", ""},
      {-1, false, "compact_stall", ""},
      {-1, false, "compact_fail", ""},
      {-1, false, "compact_success", ""},
      {-1, false, "compact_daemon_wake", ""},
      {-1, false, "compact_daemon_migrate_scanned", ""},
      {-1, false, "compact_daemon_free_scanned", ""},
      {-1, false, "htlb_buddy_alloc_success", ""},
      {-1, false, "htlb_buddy_alloc_fail", ""},
      {-1, false, "unevictable_pgs_culled", ""},
      {-1, false, "unevictable_pgs_scanned", ""},
      {-1, false, "unevictable_pgs_rescued", ""},
      {-1, false, "unevictable_pgs_mlocked", ""},
      {-1, false, "unevictable_pgs_munlocked", ""},
      {-1, false, "unevictable_pgs_cleared", ""},
      {-1, false, "unevictable_pgs_stranded", ""},
      {-1, false, "thp_fault_alloc", ""},
      {-1, false, "thp_fault_fallback", ""},
      {-1, false, "thp_collapse_alloc", ""},
      {-1, false, "thp_collapse_alloc_failed", ""},
      {-1, false, "thp_file_alloc", ""},
      {-1, false, "thp_file_mapped", ""},
      {-1, false, "thp_split_page", ""},
      {-1, false, "thp_split_page_failed", ""},
      {-1, false, "thp_deferred_split_page", ""},
      {-1, false, "thp_split_pmd", ""},
      {-1, false, "thp_split_pud", ""},
      {-1, false, "thp_zero_page_alloc", ""},
      {-1, false, "thp_zero_page_alloc_failed", ""},
      {-1, false, "thp_swpout", ""},
      {-1, false, "thp_swpout_fallback", ""},
      {-1, false, "balloon_inflate", ""},
      {-1, false, "balloon_deflate", ""},
      {-1, false, "balloon_migrate", ""},
      {-1, false, "swap_ra", ""},
      {-1, false, "swap_ra_hit", ""},
   }
)

func (d *Vmstat) probe() *[]Event {
   return &vmstatEvents
}

func (d *Vmstat) sample() []uint64 {
   buf := make([]byte, 8192)

   current := time.Now()
   delta := uint64(current.Sub(d.lastElapsed) / time.Nanosecond)
   d.lastElapsed = current

   samples := make([]uint64, len(d.enabled))
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

   for i, offset := range d.enabled {
      val := m[vmstatEvents[offset].mnemonic]
      samples[i] = (val - d.last[i]) * 1000000000 / delta
      d.last[i] = val
   }

   return samples
}

func (d *Vmstat) enable(events []uint16) {
   d.enabled = events
   d.last = make([]uint64, len(events))

   var err error
   d.file, err = os.Open("/proc/vmstat")
   validate(err)

   // update last values, discarding differences
   _ = d.sample()
}
