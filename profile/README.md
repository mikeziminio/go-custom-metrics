# Профилирование

## Запуск нагрузочного тестирования

```
task postgres-down
task postgres-up
task server-with-postgres

# в отдельном терминале - запуск k6
task test-load
```

## Исследование проблемы

```
task pprof-heap FILENAME=base
go tool pprof -top ./profile/base.pprof

Showing nodes accounting for 29282.18kB, 100% of 29282.18kB total
Showing top 10 nodes out of 67
      flat  flat%   sum%        cum   cum%
17452.48kB 59.60% 59.60% 21548.56kB 73.59%  github.com/mikeziminio/go-custom-metrics/internal/dbstorage.(*DBStorage).List
 3590.24kB 12.26% 71.86%  3590.24kB 12.26%  runtime.mallocgc
 2560.05kB  8.74% 80.60%  2560.05kB  8.74%  github.com/jackc/pgx/v5/pgtype.scanPlanString.Scan
 2086.21kB  7.12% 87.73%  2086.21kB  7.12%  bytes.growSlice
 1024.02kB  3.50% 91.23%  1024.02kB  3.50%  github.com/jackc/pgx/v5/stdlib.(*Rows).Next.func6
  521.05kB  1.78% 93.01%   521.05kB  1.78%  encoding/xml.map.init.0
```

## Исправление

После исследования становится ясно, что проблема в том, что в файле `./internal/dbstorage/dbstorage.go`
в методе `List` используется `map[string]model.Metric`. Тип был исправлен на `map[string]*model.Metric`

## Повторный замер

```
task pprof-heap FILENAME=result
go tool pprof -top ./profile/result.pprof

Showing nodes accounting for 4107.47kB, 100% of 4107.47kB total
      flat  flat%   sum%        cum   cum%
 2050.29kB 49.92% 49.92%  2050.29kB 49.92%  runtime.mallocgc
  521.05kB 12.69% 62.60%   521.05kB 12.69%  encoding/xml.map.init.0
  512.05kB 12.47% 75.07%   512.05kB 12.47%  context.(*cancelCtx).Done
  512.05kB 12.47% 87.53%   512.05kB 12.47%  github.com/jackc/pgx/v5/pgconn/ctxwatch.(*ContextWatcher).Watch.func1
  512.02kB 12.47%   100%   512.02kB 12.47%  github.com/jackc/pgx/v5/pgtype.initDefaultMap
```
