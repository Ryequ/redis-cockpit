package ops

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func ParseInfo(raw string) map[string]map[string]string {
	result := map[string]map[string]string{}
	section := "Server"
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")

	for _, line := range strings.Split(normalized, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			section = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if section == "" {
				section = "Server"
			}
			if result[section] == nil {
				result[section] = map[string]string{}
			}
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		if result[section] == nil {
			result[section] = map[string]string{}
		}
		result[section][line[:idx]] = strings.TrimSpace(line[idx+1:])
	}
	return result
}

func ParseClientList(raw string) []ClientInfo {
	var clients []ClientInfo
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := map[string]string{}
		for _, part := range strings.Fields(line) {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				fields[kv[0]] = kv[1]
			}
		}
		clients = append(clients, ClientInfo{
			ID:   fields["id"],
			Addr: fields["addr"],
			Name: fields["name"],
			DB:   fields["db"],
			Cmd:  fields["cmd"],
		})
	}
	return clients
}

func ParseConfig(reply any) map[string]string {
	result := map[string]string{}
	rv := indirect(reflect.ValueOf(reply))
	if !rv.IsValid() {
		return result
	}

	switch rv.Kind() {
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			result[fmt.Sprint(iter.Key().Interface())] = fmt.Sprint(iter.Value().Interface())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i+1 < rv.Len(); i += 2 {
			result[fmt.Sprint(rv.Index(i).Interface())] = fmt.Sprint(rv.Index(i + 1).Interface())
		}
	}
	return result
}

func ParseSlowlog(reply any) []SlowlogEntry {
	rv := indirect(reflect.ValueOf(reply))
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil
	}

	entries := make([]SlowlogEntry, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := indirect(rv.Index(i))
		if !item.IsValid() || item.Kind() != reflect.Slice && item.Kind() != reflect.Array {
			continue
		}
		if item.Len() < 4 {
			continue
		}
		entry := SlowlogEntry{
			ID:        toInt64(item.Index(0).Interface()),
			Timestamp: toInt64(item.Index(1).Interface()),
			Duration:  toInt64(item.Index(2).Interface()),
			Command:   toStrings(item.Index(3).Interface()),
		}
		if item.Len() > 4 {
			entry.Client = fmt.Sprint(item.Index(4).Interface())
		}
		if item.Len() > 5 {
			entry.Name = fmt.Sprint(item.Index(5).Interface())
		}
		entries = append(entries, entry)
	}
	return entries
}

func ParseModules(reply any) []string {
	rv := indirect(reflect.ValueOf(reply))
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil
	}
	modules := make([]string, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := indirect(rv.Index(i))
		if !item.IsValid() || item.Kind() != reflect.Map {
			continue
		}
		name := ""
		version := ""
		iter := item.MapRange()
		for iter.Next() {
			key := strings.ToLower(fmt.Sprint(iter.Key().Interface()))
			value := fmt.Sprint(iter.Value().Interface())
			switch key {
			case "name":
				name = value
			case "ver":
				version = value
			}
		}
		if name != "" {
			modules = append(modules, name+"@"+version)
		}
	}
	return modules
}

func SummarizeReplication(info map[string]string) ReplicationInfo {
	slaves, _ := strconv.Atoi(info["connected_slaves"])
	return ReplicationInfo{
		Role:            info["role"],
		MasterHost:      info["master_host"],
		MasterPort:      info["master_port"],
		ConnectedSlaves: slaves,
	}
}

func indirect(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case string:
		parsed, _ := strconv.ParseInt(v, 10, 64)
		return parsed
	default:
		return 0
	}
}

func toStrings(value any) []string {
	rv := indirect(reflect.ValueOf(value))
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil
	}
	result := make([]string, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result = append(result, fmt.Sprint(rv.Index(i).Interface()))
	}
	return result
}
