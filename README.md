# check_influxdb_query
InfluxDB query nagios check

```
usage: check_influxdb_query --warning-threshold=WARNING-THRESHOLD --critical-threshold=CRITICAL-THRESHOLD [<flags>] <query>

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
  -H, --host="localhost"       influxdb host
  -P, --port="8086"            influxdb port
  -u, --username="admin"       influxdb username
  -p, --password="admin"       influxdb password
  -d, --db="telegraf"          influxdb database name
  -w, --warning-threshold=WARNING-THRESHOLD
                               warning threshold for returned value
  -c, --critical-threshold=CRITICAL-THRESHOLD
                               critical threshold for returned value
  -o, --compare-operator="lt"  operator to compare returned value with thresholds, 'lt' or 'gt'
      --version                Show application version.

Args:
  <query>  influxdb query which returns one value to be able compare against integer thresholds
```
