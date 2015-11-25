# go-whosonfirst-breaches

THIS IS NOT (PROBABLY) READY FOR USE YET. If you're feeling dangerous by all means, dive in!

## Example

```
./bin/wof-breaches -data /usr/local/mapzen/whosonfirst-data/data /usr/local/mapzen/whosonfirst-data/meta/wof-neighbourhood-latest.csv 
[wof-breaches] 22:15:45.150146 [info] indexing /usr/local/mapzen/whosonfirst-data/meta/wof-neighbourhood-latest.csv
[wof-breaches] 22:16:34.942185 [info] indexing meta files complete
/usr/local/mapzen/whosonfirst-data/data/102/061/079/102061079.geojson
[wof-breaches] 22:17:07.033602 [info] Gowanus Heights (4201040) breaches Boerum Hill (4201040)
[wof-breaches] 22:17:07.033707 [info] Gowanus Heights (4201040) breaches Gowanus Heights (4201040)
[wof-breaches] 22:17:07.033864 [info] Gowanus Heights (4201040) breaches BoCoCa (4201040)
[wof-breaches] 22:17:07.033880 [info] time to search 3.495763ms
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-rtree
* https://github.com/akavel/polyclip-go