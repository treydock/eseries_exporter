## 1.0.0-rc.0 / 2020-11-12

### BREAKING CHANGES

* Remove --exporter.use-cache flag and all caching logic
* Remove ID related labels from all metrics as this will be instance label
* Refactor drive-statistics collector
* Refactor system-statistics collector
* CPU utilization metrics are now ratios of 0.0-1.0, add _ratio suffix to metrics
* Remove metrics
  * eseries_drive_combined_iops, eseries_drive_combined_throughput_bytes_per_second
  * eseries_drive_read_ops, eseries_drive_read_iops, eseries_drive_read_throughput_mb_per_second
  * eseries_drive_write_ops, eseries_drive_write_iops, eseries_drive_write_throughput_mb_per_second
  * eseries_system_read_ops, eseries_system_read_iops, eseries_system_read_throughput_mb_per_second
  * eseries_system_write_ops, eseries_system_write_iops, eseries_system_write_throughput_mb_per_second
  * eseries_system_cache_hit_bytes_percent, eseries_system_combined_iops
  * eseries_system_combined_throughput_mb_per_second, eseries_system_ddp_bytes_percent
  * eseries_system_full_stripe_writes_bytes_percent, eseries_system_random_ios_percent
* Rename metrics to change units of measurement
  * eseries_drive_combined_response_time_milliseconds to eseries_drive_combined_response_time_seconds
  * eseries_drive_read_response_time_milliseconds to eseries_drive_read_response_time_seconds
  * eseries_drive_write_response_time_milliseconds to eseries_drive_write_response_time_seconds
  * eseries_drive_combined_throughput_mb_per_second to eseries_drive_combined_throughput_bytes_per_second
  * eseries_system_combined_hit_response_time_milliseconds to eseries_system_combined_hit_response_time_seconds
  * eseries_system_combined_response_time_milliseconds to eseries_system_combined_response_time_seconds
  * eseries_system_read_hit_response_time_milliseconds to eseries_system_read_hit_response_time_seconds
  * eseries_system_read_response_time_milliseconds to eseries_system_read_response_time_seconds
  * eseries_system_write_hit_response_time_milliseconds to eseries_system_write_hit_response_time_seconds
  * eseries_system_write_response_time_milliseconds to eseries_system_write_response_time_seconds
* Add metrics
  * eseries_drive_idle_time_seconds_total, eseries_drive_other_ops_total, eseries_drive_other_time_seconds_total
  * eseries_drive_read_bytes_total, eseries_drive_read_ops_total, eseries_drive_read_time_seconds_total
  * eseries_drive_recovered_errors_total, eseries_drive_retried_ios_total, eseries_drive_timeouts_total
  * eseries_drive_unrecovered_errors_total
  * eseries_drive_write_bytes_total, eseries_drive_write_ops_total, eseries_drive_write_time_seconds_total
  * eseries_drive_queue_depth_total, eseries_drive_random_ios_total, eseries_drive_random_bytes_total

### Improvements

* Update to Go 1.15 and update all dependencies
* Improve status metrics to always have all possible statuses and set 1 for current status
* Add controller-statistics collector

## 0.1.1 / 2020-04-03

* Minor fix to Docker container

## 0.1.0 / 2020-04-03

* Disable drive-statistics collector by default

## 0.0.1 / 2020-04-02

* Initial Release

