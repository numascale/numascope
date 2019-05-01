package main

import (
   "os"
   "strings"
   "strconv"
   "time"
)

type Kernel struct {
   enabled     []uint16 // index into event list
   file        *os.File
   last        []uint64
   lastElapsed time.Time
}

var (
   kernelEvents = []Event{
//    include/linux/mmzone.h
      {-1, /*false,*/ "nr_free_pages", "unallocated pages"},
      {-1, /*false,*/ "nr_zone_inactive_anon", "zone inactive anonymous pages"},
      {-1, /*false,*/ "nr_zone_active_anon", "zone activate anonymous pages"},
      {-1, /*false,*/ "nr_zone_inactive_file", "zone inactive file-backed pages"},
      {-1, /*false,*/ "nr_zone_active_file", "zone active file-backed pages"},
      {-1, /*false,*/ "nr_zone_unevictable", "zone unevictable pages"},
      {-1, /*false,*/ "nr_zone_write_pending", "zone write pending pages"},
      {-1, /*false,*/ "nr_mlock", "locked pages"},
      {-1, /*false,*/ "nr_page_table_pages", "page table pages"},
      {-1, /*false,*/ "nr_kernel_stack", "kernel stack kilobytes"},
      {-1, /*false,*/ "nr_bounce", "low-memory pages allocated for DMA"},
      {-1, /*false,*/ "nr_free_cma", "free Contig Mem Alloc pages"},
      {-1, /*false,*/ "numa_hit", "allocated in intended node"},
      {-1, /*false,*/ "numa_miss", "allocated in non-intended node"},
      {-1, /*false,*/ "numa_foreign", "was intended here, hit elsewhere"},
      {-1, /*false,*/ "numa_interleave", "interleaver preferred this zone"},
      {-1, /*false,*/ "numa_local", "allocation from local node"},
      {-1, /*false,*/ "numa_other", "allocation from non-local node"},
      {-1, /*false,*/ "nr_inactive_anon", "inactive anonymous pages"},
      {-1, /*false,*/ "nr_active_anon", "active anonymous pages"},
      {-1, /*false,*/ "nr_inactive_file", "inactive file-backed pages"},
      {-1, /*false,*/ "nr_active_file", "active file-backed pages"},
      {-1, /*false,*/ "nr_unevictable", "unevictable Pages"},
      {-1, /*false,*/ "nr_slab_reclaimable", "unused bytes allocated to slab"},
      {-1, /*false,*/ "nr_slab_unreclaimable", "used bytes allocated to slab"},
      {-1, /*false,*/ "nr_isolated_anon", "temporary anonymous isolated pages"},
      {-1, /*false,*/ "nr_isolated_file", "temporary file-backed isolated pages"},
      {-1, /*false,*/ "workingset_refault", "refaults of previously evicted pages"},
      {-1, /*false,*/ "workingset_activate", "refaulted pages that were immediately activated"},
      {-1, /*false,*/ "workingset_nodereclaim", "times a shadow node has been reclaimed"},
      {-1, /*false,*/ "nr_anon_pages", "non file-backed memory-mapped pages"},
      {-1, /*false,*/ "nr_mapped", "file-backed memory-mapped pages"},
      {-1, /*false,*/ "nr_file_pages", "pagecache pages"},
      {-1, /*false,*/ "nr_dirty", "dirty pagecache pages"},
      {-1, /*false,*/ "nr_writeback", "pagecache pages pending writeback"},
      {-1, /*false,*/ "nr_writeback_temp", "pagecache pages pending writeback using temporary buffers"},
      {-1, /*false,*/ "nr_shmem", "shared memory pages including tmpfs and GEM pages"},
      {-1, /*false,*/ "nr_shmem_hugepages", "shared memory 2MB or larger pages"},
      {-1, /*false,*/ "nr_shmem_pmdmapped", "shared memory pages mapped via middle directory"},
      {-1, /*false,*/ "nr_anon_transparent_hugepages", "non file-backed 2MB or larger pages"},
      {-1, /*false,*/ "nr_unstable", "uncommitted dirty network filesystem pages"},
      {-1, /*false,*/ "nr_vmscan_write", "pages paged out"},
      {-1, /*false,*/ "nr_vmscan_immediate_reclaim", "pages ready to be reclaimed"},
      {-1, /*false,*/ "nr_dirtied", "pages dirtied"},
      {-1, /*false,*/ "nr_written", "pages written to"},
      {-1, /*false,*/ "nr_dirty_threshold", "synchronous writeback threshold bytes"},
      {-1, /*false,*/ "nr_dirty_background_threshold", "asynchronous writeback threshold bytes"},
//    include/linux/vm_event_item.h
      {-1, /*false,*/ "pgpgin", "pageins"},
      {-1, /*false,*/ "pgpgout", "pageouts"},
      {-1, /*false,*/ "pswpin", "pages swapped in"},
      {-1, /*false,*/ "pswpout", "pages swapped out"},
      {-1, /*false,*/ "pgalloc_dma32", "page allocations, DMA32 zone"},
      {-1, /*false,*/ "pgalloc_normal", "page allocations per zone, normal zone"},
      {-1, /*false,*/ "pgalloc_movable", "page allocations per zone, movable zone"},
      {-1, /*false,*/ "allocstall_dma32", "Direct reclaim calls, DMA32 zone"},
      {-1, /*false,*/ "allocstall_normal", "Direct reclaim calls, normal zone"},
      {-1, /*false,*/ "allocstall_movable", "Direct reclaim calls, movable zone"},
      {-1, /*false,*/ "pgskip_dma32", "pages unscannable, DMA32 zone"},
      {-1, /*false,*/ "pgskip_normal", "pages unscannable, normal zone"},
      {-1, /*false,*/ "pgskip_movable", "pages unscannable, movable zone"},
      {-1, /*false,*/ "pgfree", "pages freed"},
      {-1, /*false,*/ "pgactivate", "pages marked frequently used"},
      {-1, /*false,*/ "pgdeactivate", "pages marked infrequently used"},
      {-1, /*false,*/ "pglazyfree", "pages pending asynchronous freeing"},
      {-1, /*false,*/ "pgfault", "pagefaults not causing IO"},
      {-1, /*false,*/ "pgmajfault", "pagefaults causing IO"},
      {-1, /*false,*/ "pglazyfreed", "pages freed asynchronously"},
      {-1, /*false,*/ "pgrefill", "page refills"},
      {-1, /*false,*/ "pgsteal_kswapd", "page steals by kswapd"},
      {-1, /*false,*/ "pgsteal_direct", "page steals on allocation path"},
      {-1, /*false,*/ "pgscan_kswapd", "pages scanned by the kswapd daemon"},
      {-1, /*false,*/ "pgscan_direct", "pages scanned in process context"},
      {-1, /*false,*/ "pgscan_direct_throttle", "pages scanned in throttled process context"},
      {-1, /*false,*/ "zone_reclaim_failed", "reclaim failures"},
      {-1, /*false,*/ "pginodesteal", "pages reclaimed via inode freeing"},
      {-1, /*false,*/ "slabs_scanned", "slab objects scanned"},
      {-1, /*false,*/ "kswapd_inodesteal", "pages reclaimed by kswapd via inode freeing"},
      {-1, /*false,*/ "kswapd_low_wmark_hit_quickly", "times kswapd reached low watermark quickly"},
      {-1, /*false,*/ "kswapd_high_wmark_hit_quickly", "times kswapd reached high watermark quickly"},
      {-1, /*false,*/ "pageoutrun", "kswapd calls to page reclaim"},
      {-1, /*false,*/ "pgrotated", "pages reused after IO"},
      {-1, /*false,*/ "drop_pagecache", "pagecache flushes"},
      {-1, /*false,*/ "drop_slab", "slab flushes"},
      {-1, /*false,*/ "oom_kill", "out of memory kills"},
      {-1, /*false,*/ "pgmigrate_success", "pages migrated"},
      {-1, /*false,*/ "pgmigrate_fail", "pages failed migration"},
      {-1, /*false,*/ "compact_migrate_scanned", "compactable pages marked for migration in process context"},
      {-1, /*false,*/ "compact_free_scanned", "compactable free pages scanned in process context"},
      {-1, /*false,*/ "compact_isolated", "compactable pages isolated in process context"},
      {-1, /*false,*/ "compact_stall", "page compaction stalls in process context"},
      {-1, /*false,*/ "compact_fail", "page compaction failures in process context"},
      {-1, /*false,*/ "compact_success", "compaction daemon succeeded runs"},
      {-1, /*false,*/ "compact_daemon_wake", "times compaction daemon was woken"},
      {-1, /*false,*/ "compact_daemon_migrate_scanned", "pages marked for migration by compaction daemon"},
      {-1, /*false,*/ "compact_daemon_free_scanned", "free pages scanned by compaction daemon"},
      {-1, /*false,*/ "htlb_buddy_alloc_success", "2MB or larger pages allocated"},
      {-1, /*false,*/ "htlb_buddy_alloc_fail", "2MB or larger pages failed allocation"},
      {-1, /*false,*/ "unevictable_pgs_culled", "pages which became unevictable"},
      {-1, /*false,*/ "unevictable_pgs_scanned", "unevictable pages scanned"},
      {-1, /*false,*/ "unevictable_pgs_rescued", "unevictable pages became evictable"},
      {-1, /*false,*/ "unevictable_pgs_mlocked", "unevictable pages locked"},
      {-1, /*false,*/ "unevictable_pgs_munlocked", "unevictable pages unlocked"},
      {-1, /*false,*/ "unevictable_pgs_cleared", "unevictable pages zeroed"},
      {-1, /*false,*/ "unevictable_pgs_stranded", "unevictable pages which couldn't be isolated"},
      {-1, /*false,*/ "thp_fault_alloc", "2MB or larger pages allocated"},
      {-1, /*false,*/ "thp_fault_fallback", "2MB or larger pages reused"},
      {-1, /*false,*/ "thp_collapse_alloc", "2MB or larger pages from merging"},
      {-1, /*false,*/ "thp_collapse_alloc_failed", "2MB or larger page merge failure"},
      {-1, /*false,*/ "thp_file_alloc", "2MB or larger file-backed pages allocated"},
      {-1, /*false,*/ "thp_file_mapped", "2MB or larger pagefaults"},
      {-1, /*false,*/ "thp_split_page", "2MB or larger pages split to normal pages"},
      {-1, /*false,*/ "thp_split_page_failed", "2MB or larger pages split failures"},
      {-1, /*false,*/ "thp_deferred_split_page", "2MB or larger pages with deferred split"},
      {-1, /*false,*/ "thp_split_pmd", "2MB or larger pages split from middle directory"},
      {-1, /*false,*/ "thp_split_pud", "2MB or larger pages split from upper directory"},
      {-1, /*false,*/ "thp_zero_page_alloc", "2MB or larger zero pages allocated"},
      {-1, /*false,*/ "thp_zero_page_alloc_failed", "2MB or larger zero page allocation failures"},
      {-1, /*false,*/ "thp_swpout", "2MB or larger pages swapped out"},
      {-1, /*false,*/ "thp_swpout_fallback", "2MB or larger pages swapped out as normal pages"},
      {-1, /*false,*/ "balloon_inflate", "pages added to page balloon"},
      {-1, /*false,*/ "balloon_deflate", "pages removed from page balloon"},
      {-1, /*false,*/ "swap_ra", "pages swapped in due to readahead"},
      {-1, /*false,*/ "swap_ra_hit", "pages returned from swap cache due to readahead"},
   }
)

func (d *Kernel) probe() uint {
   return 1
}

func (d *Kernel) supported() *[]Event {
   return &kernelEvents
}

func (d *Kernel) sample() []uint64 {
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
      val := m[kernelEvents[offset].mnemonic]
      samples[i] = (val - d.last[i]) * 1000000000 / delta
      d.last[i] = val
   }

   return samples
}

func (d *Kernel) enable(events []uint16, discrete bool) {
   d.enabled = events
   d.last = make([]uint64, len(events))

   var err error
   d.file, err = os.Open("/proc/vmstat")
   validate(err)

   // update last values, discarding differences
   _ = d.sample()
}
