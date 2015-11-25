# go-whosonfirst-breaches

THIS IS NOT (PROBABLY) READY FOR USE YET. If you're feeling dangerous by all means, dive in!

## Example

```
./bin/wof-breaches -data /usr/local/mapzen/whosonfirst-data/data /usr/local/mapzen/whosonfirst-data/meta/wof-neighbourhood-latest.csv 
[wof-breaches] 22:48:03.508999 [info] indexing /usr/local/mapzen/whosonfirst-data/meta/wof-neighbourhood-latest.csv
[wof-breaches] 22:48:52.946219 [info] indexing meta files complete

85844047
[wof-breaches] 22:49:21.299260 [info] what records does Red Hook (/usr/local/mapzen/whosonfirst-data/data/858/440/47/85844047.geojson) breach?
[wof-breaches] 22:49:21.304940 [status] Red Hook breaches one other record in the index
[wof-breaches] 22:49:21.305178 [info] Red Hook (85844047) breaches Windsor Teraace (85865581)
[wof-breaches] 22:49:21.305198 [info] time to search 5.902957ms

85865581
[wof-breaches] 22:49:24.895528 [info] what records does Windsor Teraace (/usr/local/mapzen/whosonfirst-data/data/858/655/81/85865581.geojson) breach?
[wof-breaches] 22:49:24.900560 [status] Windsor Teraace breaches 6 other records in the index
[wof-breaches] 22:49:24.901123 [info] Windsor Teraace (85865581) breaches Sunset Park (85851575)
[wof-breaches] 22:49:24.901301 [info] Windsor Teraace (85865581) breaches Ditmas Park West (85892929)
[wof-breaches] 22:49:24.901582 [info] Windsor Teraace (85865581) breaches South Slope (85892961)
[wof-breaches] 22:49:24.901747 [info] Windsor Teraace (85865581) breaches South Brooklyn (85868663)
[wof-breaches] 22:49:24.902067 [info] Windsor Teraace (85865581) breaches Red Hook (85844047)
[wof-breaches] 22:49:24.902298 [info] Windsor Teraace (85865581) breaches Parkville (85840745)
[wof-breaches] 22:49:24.902315 [info] time to search 6.762269ms

85865903
[wof-breaches] 22:49:45.372000 [info] what records does Tenderloin (/usr/local/mapzen/whosonfirst-data/data/858/659/03/85865903.geojson) breach?
[wof-breaches] 22:49:45.377614 [status] Tenderloin does not breach any other records in the index
[wof-breaches] 22:49:45.377646 [info] time to search 5.620287ms

85887443
[wof-breaches] 22:50:16.217703 [info] what records does Mission District (/usr/local/mapzen/whosonfirst-data/data/858/874/43/85887443.geojson) breach?
[wof-breaches] 22:50:16.230024 [status] Mission District breaches 4 other records in the index
[wof-breaches] 22:50:16.230357 [info] Mission District (85887443) breaches Mission (85834637)
[wof-breaches] 22:50:16.230610 [info] Mission District (85887443) breaches Somisspo (85887457)
[wof-breaches] 22:50:16.230708 [info] Mission District (85887443) breaches La Lengua (102112179)
[wof-breaches] 22:50:16.230957 [info] Mission District (85887443) breaches Victoria Mews (85853735)
[wof-breaches] 22:50:16.230975 [info] time to search 13.246389ms
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-rtree
* https://github.com/akavel/polyclip-go