## 0.2.0 / TBD

### BREAKING CHANGES

* Remove --exporter.use-cache flag and all caching logic
* Remove ID related labels from all metrics as this will be instance label
* Remove metrics
** eseries_drive_read_ops
** eseries_drive_write_ops
** eseries_system_read_ops
** eseries_system_write_ops
* Rename metrics to change units of measurement
** eseries_drive_combined_response_time_milliseconds to eseries_drive_combined_response_time_seconds
** eseries_drive_read_response_time_milliseconds to eseries_drive_read_response_time_seconds
** eseries_drive_write_response_time_milliseconds to eseries_drive_write_response_time_seconds
** eseries_drive_combined_throughput_mb_per_second to eseries_drive_combined_throughput_bytes_per_second
** eseries_drive_read_throughput_mb_per_second to eseries_drive_read_throughput_bytes_per_second
** eseries_drive_write_throughput_mb_per_second to eseries_drive_write_throughput_bytes_per_second
** eseries_system_combined_hit_response_time_milliseconds to eseries_system_combined_hit_response_time_seconds
** eseries_system_combined_response_time_milliseconds to eseries_system_combined_response_time_seconds
** eseries_system_read_hit_response_time_milliseconds to eseries_system_read_hit_response_time_seconds
** eseries_system_read_response_time_milliseconds to eseries_system_read_response_time_seconds
** eseries_system_write_hit_response_time_milliseconds to eseries_system_write_hit_response_time_seconds
** eseries_system_write_response_time_milliseconds to eseries_system_write_response_time_seconds
** eseries_system_combined_throughput_mb_per_second to eseries_system_combined_throughput_bytes_per_second
** eseries_system_read_throughput_mb_per_second to eseries_system_read_throughput_bytes_per_second
** eseries_system_write_throughput_mb_per_second to eseries_system_write_throughput_bytes_per_second

### Improvements

* Update to Go 1.15 and update all dependencies
* Improve status metrics to always have all possible statuses and set 1 for current status

## 0.1.1 / 2020-04-03

* Minor fix to Docker container

## 0.1.0 / 2020-04-03

* Disable drive-statistics collector by default

## 0.0.1 / 2020-04-02

* Initial Release

