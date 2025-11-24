# cek kuota 
korteks@fedora ~> curl 'https://bendith.my.id/end.php?check=package&number=081938158108&version=2' \
                        -H 'accept: */*' \
                        -H 'accept-language: en-US,en;q=0.9' \
                        -b 'SITE_TOTAL_ID=d806cd6ca24a149c3059c281c55c8c62' \
                        -H 'priority: u=1, i' \
                        -H 'referer: https://bendith.my.id/' \
                        -H 'sec-ch-ua: "Chromium";v="142", "Brave";v="142", "Not_A Brand";v="99"' \
                        -H 'sec-ch-ua-mobile: ?0' \
                        -H 'sec-ch-ua-platform: "Linux"' \
                        -H 'sec-fetch-dest: empty' \
                        -H 'sec-fetch-mode: cors' \
                        -H 'sec-fetch-site: same-origin' \
                        -H 'sec-gpc: 1' \
                        -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36'
{"success":true,"code":"000","message":"","data":{"subs_info":{"msisdn":"6281938158108","operator":"XL","id_verified":"Sudah","net_type":"4G","tenure":"0 Tahun 2 Bulan","exp_date":"29-11-2025","grace_until":"29-12-2025","volte":{"device":true,"area":true,"simcard":true}},"package_info":{"error_message":null,"packages":[{"name":"Paket Swadaya XLalu Untukmu 7D 20K","expiry":"28-11-2025","timestamp":1764349199,"quotas":[{"name":"24jam di 2G3G4G","percent":0,"total":"4294967296","remaining":"0"},{"name":"Aplikasi Gojek dan Waze","percent":99.68,"total":"500MB","remaining":"498MB"},{"name":"Akses Aplikasi Grab","percent":0,"total":"1KB","remaining":"0KB"},{"name":"Nelp (ke XL/Axis)","percent":100,"total":"50000 Menit","remaining":"50000 Menit"},{"name":"Nelp Oprt Lain","percent":100,"total":"15 Menit","remaining":"15 Menit"},{"name":"SMS (ke XL/Axis)","percent":100,"total":"200000 SMS","remaining":"200000 SMS"}]},{"name":"Bonus 2GB 3hr","expiry":"24-11-2025","timestamp":1764003599,"quotas":[{"name":"24jam di 2G3G4G","percent":0,"total":"2GB","remaining":"0KB"}]}]}}}⏎                              


# cek area
L adalah Area

curl 'https://bendith.my.id/area_list.json' \
  -H 'sec-ch-ua-platform: "Linux"' \
  -H 'Referer: https://bendith.my.id/' \
  -H 'User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36' \
  -H 'sec-ch-ua: "Chromium";v="142", "Brave";v="142", "Not_A Brand";v="99"' \
  -H 'sec-ch-ua-mobile: ?0'

{
  "akrab": [
    {
      "value": "L4",
      "label": "Kab. Aceh Barat"
    },
    {
      "value": "L2",
      "label": "Kab. Aceh Barat Daya"
    },
    {
      "value": "L2",
      "label": "Kab. Aceh Besar"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Jaya"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Aceh Singkil"
    },
    {
      "value": "L4",
      "label": "Kab. Aceh Tamiang"
    },
    {
      "value": "L4",
      "label": "Kab. Aceh Tengah"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Tenggara"
    },
    {
      "value": "L4",
      "label": "Kab. Aceh Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Aceh Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Agam"
    },
    {
      "value": "L4",
      "label": "Kab. Alor"
    },
    {
      "value": "L3",
      "label": "Kab. Asahan"
    },
    {
      "value": "L1",
      "label": "Kab. Asmat"
    },
    {
      "value": "L2",
      "label": "Kab. Badung"
    },
    {
      "value": "L3",
      "label": "Kab. Balangan"
    },
    {
      "value": "L1",
      "label": "Kab. Bandung"
    },
    {
      "value": "L2",
      "label": "Kab. Bandung Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Banggai"
    },
    {
      "value": "L3",
      "label": "Kab. Banggai Kepulauan"
    },
    {
      "value": "L3",
      "label": "Kab. Banggai Laut"
    },
    {
      "value": "L3",
      "label": "Kab. Bangka"
    },
    {
      "value": "L3",
      "label": "Kab. Bangka Barat"
    },
    {
      "value": "L2",
      "label": "Kab. Bangka Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Bangka Tengah"
    },
    {
      "value": "L2",
      "label": "Kab. Bangkalan"
    },
    {
      "value": "L2",
      "label": "Kab. Bangli"
    },
    {
      "value": "L2",
      "label": "Kab. Banjar"
    },
    {
      "value": "L4",
      "label": "Kab. Banjarnegara"
    },
    {
      "value": "L1",
      "label": "Kab. Bantul"
    },
    {
      "value": "L3",
      "label": "Kab. Banyuasin"
    },
    {
      "value": "L4",
      "label": "Kab. Banyumas"
    },
    {
      "value": "L2",
      "label": "Kab. Banyuwangi"
    },
    {
      "value": "L3",
      "label": "Kab. Barito Kuala"
    },
    {
      "value": "L4",
      "label": "Kab. Barito Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Barito Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Barito Utara"
    },
    {
      "value": "L2",
      "label": "Kab. Barru"
    },
    {
      "value": "L4",
      "label": "Kab. Batang"
    },
    {
      "value": "L3",
      "label": "Kab. Batanghari"
    },
    {
      "value": "L3",
      "label": "Kab. Batu Bara"
    },
    {
      "value": "L3",
      "label": "Kab. Bekasi"
    },
    {
      "value": "L2",
      "label": "Kab. Belitung"
    },
    {
      "value": "L2",
      "label": "Kab. Belitung Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Belu"
    },
    {
      "value": "L4",
      "label": "Kab. Bener Meriah"
    },
    {
      "value": "L3",
      "label": "Kab. Bengkalis"
    },
    {
      "value": "L4",
      "label": "Kab. Bengkayang"
    },
    {
      "value": "L4",
      "label": "Kab. Bengkulu Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Bengkulu Tengah"
    },
    {
      "value": "L4",
      "label": "Kab. Bengkulu Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Berau"
    },
    {
      "value": "L1",
      "label": "Kab. Biak Numfor"
    },
    {
      "value": "L4",
      "label": "Kab. Bima"
    },
    {
      "value": "L3",
      "label": "Kab. Bintan"
    },
    {
      "value": "L4",
      "label": "Kab. Bireuen"
    },
    {
      "value": "L4",
      "label": "Kab. Blitar"
    },
    {
      "value": "L4",
      "label": "Kab. Blora"
    },
    {
      "value": "L4",
      "label": "Kab. Boalemo"
    },
    {
      "value": "L3",
      "label": "Kab. Bogor"
    },
    {
      "value": "L4",
      "label": "Kab. Bojonegoro"
    },
    {
      "value": "L4",
      "label": "Kab. Bolaang Mongondow"
    },
    {
      "value": "L4",
      "label": "Kab. Bolaang Mongondow Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Bolaang Mongondow Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Bolaang Mongondow Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Bombana"
    },
    {
      "value": "L4",
      "label": "Kab. Bondowoso"
    },
    {
      "value": "L4",
      "label": "Kab. Bone"
    },
    {
      "value": "L4",
      "label": "Kab. Bone Bolango"
    },
    {
      "value": "L1",
      "label": "Kab. Boven Digoel"
    },
    {
      "value": "L3",
      "label": "Kab. Boyolali"
    },
    {
      "value": "L2",
      "label": "Kab. Brebes"
    },
    {
      "value": "L1",
      "label": "Kab. Buleleng"
    },
    {
      "value": "L4",
      "label": "Kab. Bulukumba"
    },
    {
      "value": "L4",
      "label": "Kab. Bulungan"
    },
    {
      "value": "L4",
      "label": "Kab. Bungo"
    },
    {
      "value": "L4",
      "label": "Kab. Buol"
    },
    {
      "value": "L1",
      "label": "Kab. Buru"
    },
    {
      "value": "L1",
      "label": "Kab. Buru Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Buton"
    },
    {
      "value": "L3",
      "label": "Kab. Buton Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Buton Tengah"
    },
    {
      "value": "L3",
      "label": "Kab. Ciamis"
    },
    {
      "value": "L4",
      "label": "Kab. Cianjur"
    },
    {
      "value": "L3",
      "label": "Kab. Cilacap"
    },
    {
      "value": "L2",
      "label": "Kab. Cirebon"
    },
    {
      "value": "L2",
      "label": "Kab. Dairi"
    },
    {
      "value": "L3",
      "label": "Kab. Deli Serdang"
    },
    {
      "value": "L4",
      "label": "Kab. Demak"
    },
    {
      "value": "L4",
      "label": "Kab. Dharmasraya"
    },
    {
      "value": "L2",
      "label": "Kab. Dompu"
    },
    {
      "value": "L4",
      "label": "Kab. Donggala"
    },
    {
      "value": "L4",
      "label": "Kab. Empat Lawang"
    },
    {
      "value": "L4",
      "label": "Kab. Ende"
    },
    {
      "value": "L3",
      "label": "Kab. Enrekang"
    },
    {
      "value": "L1",
      "label": "Kab. Fak Fak"
    },
    {
      "value": "L4",
      "label": "Kab. Flores Timur"
    },
    {
      "value": "L3",
      "label": "Kab. Garut"
    },
    {
      "value": "L2",
      "label": "Kab. Gayo Lues"
    },
    {
      "value": "L2",
      "label": "Kab. Gianyar"
    },
    {
      "value": "L4",
      "label": "Kab. Gorontalo"
    },
    {
      "value": "L4",
      "label": "Kab. Gorontalo Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Gowa"
    },
    {
      "value": "L4",
      "label": "Kab. Gresik"
    },
    {
      "value": "L3",
      "label": "Kab. Grobogan"
    },
    {
      "value": "L4",
      "label": "Kab. Gunung Mas"
    },
    {
      "value": "L1",
      "label": "Kab. Gunungkidul"
    },
    {
      "value": "L1",
      "label": "Kab. Halmahera Barat"
    },
    {
      "value": "L1",
      "label": "Kab. Halmahera Selatan"
    },
    {
      "value": "L1",
      "label": "Kab. Halmahera Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Hulu Sungai Selatan"
    },
    {
      "value": "L2",
      "label": "Kab. Hulu Sungai Tengah"
    },
    {
      "value": "L2",
      "label": "Kab. Hulu Sungai Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Humbang Hasundutan"
    },
    {
      "value": "L4",
      "label": "Kab. Indragiri Hilir"
    },
    {
      "value": "L3",
      "label": "Kab. Indragiri Hulu"
    },
    {
      "value": "L2",
      "label": "Kab. Indramayu"
    },
    {
      "value": "L1",
      "label": "Kab. Jayapura"
    },
    {
      "value": "L4",
      "label": "Kab. Jember"
    },
    {
      "value": "L1",
      "label": "Kab. Jembrana"
    },
    {
      "value": "L4",
      "label": "Kab. Jeneponto"
    },
    {
      "value": "L4",
      "label": "Kab. Jepara"
    },
    {
      "value": "L4",
      "label": "Kab. Jombang"
    },
    {
      "value": "L3",
      "label": "Kab. Kampar"
    },
    {
      "value": "L3",
      "label": "Kab. Kapuas"
    },
    {
      "value": "L4",
      "label": "Kab. Kapuas Hulu"
    },
    {
      "value": "L4",
      "label": "Kab. Karanganyar"
    },
    {
      "value": "L2",
      "label": "Kab. Karangasem"
    },
    {
      "value": "L4",
      "label": "Kab. Karawang"
    },
    {
      "value": "L3",
      "label": "Kab. Karimun"
    },
    {
      "value": "L2",
      "label": "Kab. Karo"
    },
    {
      "value": "L4",
      "label": "Kab. Katingan"
    },
    {
      "value": "L4",
      "label": "Kab. Kaur"
    },
    { "value": "L4", "label": "Kab. Kayong Utara" },
    {
      "value": "L2",
      "label": "Kab. Kebumen"
    },
    {
      "value": "L4",
      "label": "Kab. Kediri"
    },
    {
      "value": "L1",
      "label": "Kab. Keerom"
    },
    {
      "value": "L3",
      "label": "Kab. Kendal"
    },
    {
      "value": "L4",
      "label": "Kab. Kepahiang"
    },
    {
      "value": "L4",
      "label": "Kab. Kepulauan Anambas"
    },
    {
      "value": "L1",
      "label": "Kab. Kepulauan Aru"
    },
    {
      "value": "L3",
      "label": "Kab. Kepulauan Meranti"
    },
    {
      "value": "L4",
      "label": "Kab. Kepulauan Sangihe"
    },
    {
      "value": "L4",
      "label": "Kab. Kepulauan Selayar"
    },
    {
      "value": "L3",
      "label": "Kab. Kepulauan Seribu"
    },
    {
      "value": "L1",
      "label": "Kab. Kepulauan Sula"
    },
    {
      "value": "L4",
      "label": "Kab. Kepulauan Talaud"
    },
    {
      "value": "L1",
      "label": "Kab. Kepulauan Yapen"
    },
    {
      "value": "L4",
      "label": "Kab. Kerinci"
    },
    {
      "value": "L4",
      "label": "Kab. Ketapang"
    },
    {
      "value": "L4",
      "label": "Kab. Klaten"
    },
    {
      "value": "L2",
      "label": "Kab. Klungkung"
    },
    {
      "value": "L4",
      "label": "Kab. Kolaka"
    },
    {
      "value": "L4",
      "label": "Kab. Kolaka Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Kolaka Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Konawe"
    },
    {
      "value": "L2",
      "label": "Kab. Konawe Kepulauan"
    },
    {
      "value": "L4",
      "label": "Kab. Konawe Selatan"
    },
    {
      "value": "L2",
      "label": "Kab. Konawe Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Kotabaru"
    },
    {
      "value": "L4",
      "label": "Kab. Kotawaringin Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Kotawaringin Timur"
    },
    {
      "value": "L3",
      "label": "Kab. Kuantan Singingi"
    },
    {
      "value": "L4",
      "label": "Kab. Kubu Raya"
    },
    {
      "value": "L4",
      "label": "Kab. Kudus"
    },
    {
      "value": "L1",
      "label": "Kab. Kulon Progo"
    },
    {
      "value": "L1",
      "label": "Kab. Kuningan"
    },
    {
      "value": "L4",
      "label": "Kab. Kupang"
    },
    {
      "value": "L4",
      "label": "Kab. Kutai Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Kutai Kartanegara"
    },
    {
      "value": "L4",
      "label": "Kab. Kutai Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Labuhanbatu"
    },
    {
      "value": "L4",
      "label": "Kab. Labuhanbatu Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Labuhanbatu Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Lahat"
    },
    {
      "value": "L4",
      "label": "Kab. Lamandau"
    },
    {
      "value": "L4",
      "label": "Kab. Lamongan"
    },
    {
      "value": "L4",
      "label": "Kab. Lampung Barat"
    },
    {
      "value": "L3",
      "label": "Kab. Lampung Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Lampung Tengah"
    },
    {
      "value": "L4",
      "label": "Kab. Lampung Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Lampung Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Landak"
    },
    {
      "value": "L3",
      "label": "Kab. Langkat"
    },
    {
      "value": "L2",
      "label": "Kab. Lebak"
    },
    {
      "value": "L4",
      "label": "Kab. Lebong"
    },
    {
      "value": "L4",
      "label": "Kab. Lembata"
    },
    {
      "value": "L4",
      "label": "Kab. Lima Puluh Kota"
    },
    {
      "value": "L4",
      "label": "Kab. Lingga"
    },
    {
      "value": "L2",
      "label": "Kab. Lombok Barat"
    },
    {
      "value": "L2",
      "label": "Kab. Lombok Tengah"
    },
    {
      "value": "L2",
      "label": "Kab. Lombok Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Lombok Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Lumajang"
    },
    {
      "value": "L4",
      "label": "Kab. Luwu"
    },
    {
      "value": "L3",
      "label": "Kab. Luwu Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Luwu Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Madiun"
    },
    {
      "value": "L4",
      "label": "Kab. Magelang"
    },
    {
      "value": "L4",
      "label": "Kab. Magetan"
    },
    {
      "value": "L4",
      "label": "Kab. Mahakam Ulu"
    },
    {
      "value": "L3",
      "label": "Kab. Majalengka"
    },
    {
      "value": "L3",
      "label": "Kab. Majene"
    },
    {
      "value": "L4",
      "label": "Kab. Malaka"
    },
    {
      "value": "L4",
      "label": "Kab. Malang"
    },
    {
      "value": "L4",
      "label": "Kab. Malinau"
    },
    {
      "value": "L1",
      "label": "Kab. Maluku Barat Daya"
    },
    {
      "value": "L1",
      "label": "Kab. Maluku Tengah"
    },
    {
      "value": "L1",
      "label": "Kab. Maluku Tenggara Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Mamasa"
    },
    {
      "value": "L4",
      "label": "Kab. Mamuju"
    },
    {
      "value": "L3",
      "label": "Kab. Mamuju Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Mandailing Natal"
    },
    {
      "value": "L4",
      "label": "Kab. Manggarai"
    },
    {
      "value": "L4",
      "label": "Kab. Manggarai Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Manggarai Timur"
    },
    {
      "value": "L1",
      "label": "Kab. Manokwari"
    },
    {
      "value": "L1",
      "label": "Kab. Manokwari Selatan"
    },
    {
      "value": "L1",
      "label": "Kab. Mappi"
    },
    {
      "value": "L4",
      "label": "Kab. Maros"
    },
    {
      "value": "L1",
      "label": "Kab. Maybrat"
    },
    {
      "value": "L4",
      "label": "Kab. Melawi"
    },
    {
      "value": "L4",
      "label": "Kab. Mempawah"
    },
    {
      "value": "L4",
      "label": "Kab. Merangin"
    },
    {
      "value": "L1",
      "label": "Kab. Merauke"
    },
    {
      "value": "L4",
      "label": "Kab. Mesuji"
    },
    {
      "value": "L1",
      "label": "Kab. Mimika"
    },
    {
      "value": "L4",
      "label": "Kab. Minahasa"
    },
    {
      "value": "L4",
      "label": "Kab. Minahasa Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Minahasa Tenggara"
    },
    {
      "value": "L4",
      "label": "Kab. Minahasa Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Mojokerto"
    },
    {
      "value": "L3",
      "label": "Kab. Morowali"
    },
    {
      "value": "L3",
      "label": "Kab. Morowali Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Muara Enim"
    },
    {
      "value": "L3",
      "label": "Kab. Muaro Jambi"
    },
    {
      "value": "L4",
      "label": "Kab. Muko Muko"
    },
    {
      "value": "L3",
      "label": "Kab. Muna"
    },
    {
      "value": "L3",
      "label": "Kab. Muna Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Murung Raya"
    },
    {
      "value": "L4",
      "label": "Kab. Musi Banyuasin"
    },
    {
      "value": "L4",
      "label": "Kab. Musi Rawas"
    },
    {
      "value": "L4",
      "label": "Kab. Musi Rawas Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Nagan Raya"
    },
    {
      "value": "L4",
      "label": "Kab. Nagekeo"
    },
    {
      "value": "L4",
      "label": "Kab. Natuna"
    },
    {
      "value": "L4",
      "label": "Kab. Ngada"
    },
    {
      "value": "L4",
      "label": "Kab. Nganjuk"
    },
    {
      "value": "L4",
      "label": "Kab. Ngawi"
    },
    {
      "value": "L4",
      "label": "Kab. Nias"
    },
    {
      "value": "L4",
      "label": "Kab. Nunukan"
    },
    {
      "value": "L3",
      "label": "Kab. Ogan Ilir"
    },
    {
      "value": "L3",
      "label": "Kab. Ogan Komering Ilir"
    },
    {
      "value": "L4",
      "label": "Kab. Ogan Komering Ulu"
    },
    {
      "value": "L4",
      "label": "Kab. Ogan Komering Ulu Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Ogan Komering Ulu Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Pacitan"
    },
    {
      "value": "L4",
      "label": "Kab. Padang Lawas"
    },
    {
      "value": "L4",
      "label": "Kab. Padang Lawas Utara"
    },
    { "value": "L3", "label": "Kab. Padang Pariaman" },
    {
      "value": "L4",
      "label": "Kab. Pahuwato"
    },
    {
      "value": "L4",
      "label": "Kab. Pakpak Bharat"
    },
    {
      "value": "L2",
      "label": "Kab. Pamekasan"
    },
    {
      "value": "L2",
      "label": "Kab. Pandeglang"
    },
    {
      "value": "L4",
      "label": "Kab. Pangandaran"
    },
    {
      "value": "L4",
      "label": "Kab. Pangkajene Kepulauan"
    },
    {
      "value": "L4",
      "label": "Kab. Parigi Moutong"
    },
    {
      "value": "L4",
      "label": "Kab. Pasaman"
    },
    {
      "value": "L4",
      "label": "Kab. Pasaman Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Paser"
    },
    {
      "value": "L4",
      "label": "Kab. Pasuruan"
    },
    {
      "value": "L4",
      "label": "Kab. Pati"
    },
    {
      "value": "L4",
      "label": "Kab. Pekalongan"
    },
    {
      "value": "L3",
      "label": "Kab. Pelalawan"
    },
    {
      "value": "L2",
      "label": "Kab. Pemalang"
    },
    {
      "value": "L4",
      "label": "Kab. Penajam Paser Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Penukal Abab Lematang Ilir"
    },
    {
      "value": "L3",
      "label": "Kab. Pesawaran"
    },
    {
      "value": "L4",
      "label": "Kab. Pesisir Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Pesisir Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Pidie"
    },
    {
      "value": "L4",
      "label": "Kab. Pidie Jaya"
    },
    {
      "value": "L2",
      "label": "Kab. Pinrang"
    },
    {
      "value": "L3",
      "label": "Kab. Polewali Mandar"
    },
    {
      "value": "L4",
      "label": "Kab. Ponorogo"
    },
    {
      "value": "L4",
      "label": "Kab. Poso"
    },
    {
      "value": "L3",
      "label": "Kab. Pringsewu"
    },
    {
      "value": "L3",
      "label": "Kab. Probolinggo"
    },
    {
      "value": "L3",
      "label": "Kab. Pulang Pisau"
    },
    {
      "value": "L1",
      "label": "Kab. Pulau Taliabu"
    },
    {
      "value": "L4",
      "label": "Kab. Purbalingga"
    },
    {
      "value": "L1",
      "label": "Kab. Purwakarta"
    },
    {
      "value": "L4",
      "label": "Kab. Purworejo"
    },
    {
      "value": "L4",
      "label": "Kab. Rejang Lebong"
    },
    {
      "value": "L3",
      "label": "Kab. Rembang"
    },
    {
      "value": "L3",
      "label": "Kab. Rokan Hilir"
    },
    {
      "value": "L4",
      "label": "Kab. Rokan Hulu"
    },
    {
      "value": "L4",
      "label": "Kab. Rote Ndao"
    },
    {
      "value": "L4",
      "label": "Kab. Sabu Raijua"
    },
    {
      "value": "L4",
      "label": "Kab. Sambas"
    },
    {
      "value": "L4",
      "label": "Kab. Samosir"
    },
    {
      "value": "L2",
      "label": "Kab. Sampang"
    },
    {
      "value": "L4",
      "label": "Kab. Sanggau"
    },
    {
      "value": "L3",
      "label": "Kab. Sarolangun"
    },
    {
      "value": "L4",
      "label": "Kab. Sekadau"
    },
    {
      "value": "L3",
      "label": "Kab. Seluma"
    },
    {
      "value": "L3",
      "label": "Kab. Semarang"
    },
    {
      "value": "L1",
      "label": "Kab. Seram Bagian Barat"
    },
    {
      "value": "L1",
      "label": "Kab. Seram Bagian Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Serang"
    },
    {
      "value": "L3",
      "label": "Kab. Serdang Bedagai"
    },
    {
      "value": "L4",
      "label": "Kab. Seruyan"
    },
    {
      "value": "L3",
      "label": "Kab. Siak"
    },
    {
      "value": "L4",
      "label": "Kab. Siau Tagulandang Biaro"
    },
    {
      "value": "L3",
      "label": "Kab. Sidenreng Rappang"
    },
    {
      "value": "L2",
      "label": "Kab. Sidoarjo"
    },
    {
      "value": "L4",
      "label": "Kab. Sigi"
    },
    {
      "value": "L3",
      "label": "Kab. Sijunjung"
    },
    {
      "value": "L4",
      "label": "Kab. Sikka"
    },
    {
      "value": "L4",
      "label": "Kab. Simalungun"
    },
    {
      "value": "L4",
      "label": "Kab. Simeulue"
    },
    {
      "value": "L3",
      "label": "Kab. Sinjai"
    },
    {
      "value": "L4",
      "label": "Kab. Sintang"
    },
    {
      "value": "L4",
      "label": "Kab. Situbondo"
    },
    {
      "value": "L1",
      "label": "Kab. Sleman"
    },
    {
      "value": "L4",
      "label": "Kab. Solok"
    },
    {
      "value": "L3",
      "label": "Kab. Solok Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Soppeng"
    },
    {
      "value": "L1",
      "label": "Kab. Sorong"
    },
    {
      "value": "L1",
      "label": "Kab. Sorong Selatan"
    },
    { "value": "L4", "label": "Kab. Sragen" },
    {
      "value": "L2",
      "label": "Kab. Subang"
    },
    {
      "value": "L4",
      "label": "Kab. Sukabumi"
    },
    {
      "value": "L4",
      "label": "Kab. Sukamara"
    },
    {
      "value": "L4",
      "label": "Kab. Sukoharjo"
    },
    {
      "value": "L4",
      "label": "Kab. Sumba Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Sumba Barat Daya"
    },
    {
      "value": "L4",
      "label": "Kab. Sumba Tengah"
    },
    {
      "value": "L4",
      "label": "Kab. Sumba Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Sumbawa"
    },
    {
      "value": "L2",
      "label": "Kab. Sumbawa Barat"
    },
    {
      "value": "L3",
      "label": "Kab. Sumedang"
    },
    {
      "value": "L2",
      "label": "Kab. Sumenep"
    },
    {
      "value": "L2",
      "label": "Kab. Tabalong"
    },
    {
      "value": "L2",
      "label": "Kab. Tabanan"
    },
    {
      "value": "L4",
      "label": "Kab. Takalar"
    },
    {
      "value": "L4",
      "label": "Kab. Tana Tidung"
    },
    {
      "value": "L3",
      "label": "Kab. Tana Toraja"
    },
    {
      "value": "L2",
      "label": "Kab. Tanah Bumbu"
    },
    {
      "value": "L4",
      "label": "Kab. Tanah Datar"
    },
    {
      "value": "L3",
      "label": "Kab. Tanah Laut"
    },
    {
      "value": "L2",
      "label": "Kab. Tangerang"
    },
    {
      "value": "L4",
      "label": "Kab. Tanggamus"
    },
    {
      "value": "L3",
      "label": "Kab. Tanjung Jabung Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Tanjung Jabung Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Tapanuli Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Tapanuli Tengah"
    },
    {
      "value": "L4",
      "label": "Kab. Tapanuli Utara"
    },
    {
      "value": "L2",
      "label": "Kab. Tapin"
    },
    {
      "value": "L3",
      "label": "Kab. Tasikmalaya"
    },
    {
      "value": "L4",
      "label": "Kab. Tebo"
    },
    {
      "value": "L2",
      "label": "Kab. Tegal"
    },
    {
      "value": "L1",
      "label": "Kab. Teluk Bintuni"
    },
    {
      "value": "L1",
      "label": "Kab. Teluk Wondama"
    },
    {
      "value": "L4",
      "label": "Kab. Temanggung"
    },
    {
      "value": "L4",
      "label": "Kab. Timor Tengah Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Timor Tengah Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Toba Samosir"
    },
    {
      "value": "L3",
      "label": "Kab. Tojo Una Una"
    },
    {
      "value": "L3",
      "label": "Kab. Toli Toli"
    },
    {
      "value": "L4",
      "label": "Kab. Toraja Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Trenggalek"
    },
    {
      "value": "L4",
      "label": "Kab. Tuban"
    },
    {
      "value": "L4",
      "label": "Kab. Tulang Bawang"
    },
    {
      "value": "L4",
      "label": "Kab. Tulang Bawang Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Tulungagung"
    },
    {
      "value": "L4",
      "label": "Kab. Wajo"
    },
    {
      "value": "L4",
      "label": "Kab. Way Kanan"
    },
    {
      "value": "L4",
      "label": "Kab. Wonogiri"
    },
    {
      "value": "L4",
      "label": "Kab. Wonosobo"
    },
    {
      "value": "L1",
      "label": "Kota Ambon"
    },
    {
      "value": "L4",
      "label": "Kota Balikpapan"
    },
    {
      "value": "L3",
      "label": "Kota Banda Aceh"
    },
    {
      "value": "L3",
      "label": "Kota Bandar Lampung"
    },
    {
      "value": "L1",
      "label": "Kota Bandung"
    },
    {
      "value": "L3",
      "label": "Kota Banjar"
    },
    {
      "value": "L2",
      "label": "Kota Banjarbaru"
    },
    {
      "value": "L2",
      "label": "Kota Banjarmasin"
    },
    {
      "value": "L2",
      "label": "Kota Batam"
    },
    {
      "value": "L4",
      "label": "Kota Batu"
    },
    {
      "value": "L3",
      "label": "Kota Bau Bau"
    },
    {
      "value": "L3",
      "label": "Kota Bekasi"
    },
    {
      "value": "L4",
      "label": "Kota Bengkulu"
    },
    {
      "value": "L2",
      "label": "Kota Bima"
    },
    {
      "value": "L3",
      "label": "Kota Binjai"
    },
    {
      "value": "L4",
      "label": "Kota Bitung"
    },
    { "value": "L4", "label": "Kota Blitar" },
    {
      "value": "L3",
      "label": "Kota Bogor"
    },
    {
      "value": "L4",
      "label": "Kota Bontang"
    },
    {
      "value": "L4",
      "label": "Kota Bukittinggi"
    },
    {
      "value": "L2",
      "label": "Kota Cilegon"
    },
    {
      "value": "L2",
      "label": "Kota Cimahi"
    },
    {
      "value": "L2",
      "label": "Kota Cirebon"
    },
    {
      "value": "L2",
      "label": "Kota Denpasar"
    },
    {
      "value": "L3",
      "label": "Kota Depok"
    },
    {
      "value": "L3",
      "label": "Kota Dumai"
    },
    {
      "value": "L4",
      "label": "Kota Gorontalo"
    },
    {
      "value": "L3",
      "label": "Kota Gunungsitoli"
    },
    {
      "value": "L2",
      "label": "Kota Jakarta Barat"
    },
    {
      "value": "L2",
      "label": "Kota Jakarta Pusat"
    },
    {
      "value": "L2",
      "label": "Kota Jakarta Selatan"
    },
    {
      "value": "L2",
      "label": "Kota Jakarta Timur"
    },
    {
      "value": "L2",
      "label": "Kota Jakarta Utara"
    },
    {
      "value": "L3",
      "label": "Kota Jambi"
    },
    {
      "value": "L1",
      "label": "Kota Jayapura"
    },
    {
      "value": "L4",
      "label": "Kota Kediri"
    },
    {
      "value": "L4",
      "label": "Kota Kendari"
    },
    {
      "value": "L4",
      "label": "Kota Kotamobagu"
    },
    {
      "value": "L4",
      "label": "Kota Kupang"
    },
    {
      "value": "L4",
      "label": "Kota Langsa"
    },
    {
      "value": "L4",
      "label": "Kota Lhokseumawe"
    },
    {
      "value": "L4",
      "label": "Kota Lubuk Linggau"
    },
    {
      "value": "L4",
      "label": "Kota Madiun"
    },
    {
      "value": "L4",
      "label": "Kota Magelang"
    },
    {
      "value": "L4",
      "label": "Kota Makassar"
    },
    {
      "value": "L4",
      "label": "Kota Malang"
    },
    {
      "value": "L4",
      "label": "Kota Manado"
    },
    {
      "value": "L2",
      "label": "Kota Mataram"
    },
    {
      "value": "L2",
      "label": "Kota Medan"
    },
    {
      "value": "L3",
      "label": "Kota Metro"
    },
    {
      "value": "L4",
      "label": "Kota Mojokerto"
    },
    {
      "value": "L3",
      "label": "Kota Padang"
    },
    {
      "value": "L3",
      "label": "Kota Padang Panjang"
    },
    {
      "value": "L4",
      "label": "Kota Padangsidimpuan"
    },
    {
      "value": "L4",
      "label": "Kota Pagar Alam"
    },
    {
      "value": "L3",
      "label": "Kota Palangkaraya"
    },
    {
      "value": "L3",
      "label": "Kota Palembang"
    },
    {
      "value": "L4",
      "label": "Kota Palopo"
    },
    {
      "value": "L3",
      "label": "Kota Palu"
    },
    {
      "value": "L2",
      "label": "Kota Pangkal Pinang"
    },
    {
      "value": "L2",
      "label": "Kota Pare Pare"
    },
    {
      "value": "L4",
      "label": "Kota Pariaman"
    },
    {
      "value": "L4",
      "label": "Kota Pasuruan"
    },
    {
      "value": "L3",
      "label": "Kota Payakumbuh"
    },
    {
      "value": "L4",
      "label": "Kota Pekalongan"
    },
    {
      "value": "L2",
      "label": "Kota Pekanbaru"
    },
    {
      "value": "L4",
      "label": "Kota Pematangsiantar"
    },
    {
      "value": "L4",
      "label": "Kota Pontianak"
    },
    {
      "value": "L4",
      "label": "Kota Prabumulih"
    },
    { "value": "L2", "label": "Kota Probolinggo" },
    {
      "value": "L2",
      "label": "Kota Sabang"
    },
    {
      "value": "L3",
      "label": "Kota Salatiga"
    },
    {
      "value": "L4",
      "label": "Kota Samarinda"
    },
    {
      "value": "L4",
      "label": "Kota Sawahlunto"
    },
    {
      "value": "L2",
      "label": "Kota Semarang"
    },
    {
      "value": "L2",
      "label": "Kota Serang"
    },
    {
      "value": "L4",
      "label": "Kota Sibolga"
    },
    {
      "value": "L4",
      "label": "Kota Singkawang"
    },
    {
      "value": "L4",
      "label": "Kota Solok"
    },
    {
      "value": "L1",
      "label": "Kota Sorong"
    },
    {
      "value": "L4",
      "label": "Kota Subulussalam"
    },
    {
      "value": "L4",
      "label": "Kota Sukabumi"
    },
    {
      "value": "L4",
      "label": "Kota Sungai Penuh"
    },
    {
      "value": "L2",
      "label": "Kota Surabaya"
    },
    {
      "value": "L2",
      "label": "Kota Surakarta"
    },
    {
      "value": "L2",
      "label": "Kota Tangerang"
    },
    {
      "value": "L2",
      "label": "Kota Tangerang Selatan"
    },
    {
      "value": "L3",
      "label": "Kota Tanjung Balai"
    },
    {
      "value": "L3",
      "label": "Kota Tanjung Pinang"
    },
    {
      "value": "L4",
      "label": "Kota Tarakan"
    },
    {
      "value": "L3",
      "label": "Kota Tasikmalaya"
    },
    {
      "value": "L3",
      "label": "Kota Tebing Tinggi"
    },
    {
      "value": "L2",
      "label": "Kota Tegal"
    },
    {
      "value": "L1",
      "label": "Kota Tidore Kepulauan"
    },
    {
      "value": "L4",
      "label": "Kota Tomohon"
    },
    {
      "value": "L1",
      "label": "Kota Yogyakarta"
    }
  ],
  "bepu": [
    {
      "value": "L4",
      "label": "Kab. Aceh Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Aceh Barat Daya"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Besar"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Jaya"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Aceh Singkil"
    },
    {
      "value": "L4",
      "label": "Kab. Aceh Tamiang"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Tengah"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Tenggara"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Timur"
    },
    {
      "value": "L3",
      "label": "Kab. Aceh Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Agam"
    },
    {
      "value": "L4",
      "label": "Kab. Alor"
    },
    {
      "value": "L2",
      "label": "Kab. Asahan"
    },
    {
      "value": "L1",
      "label": "Kab. Asmat"
    },
    {
      "value": "L2",
      "label": "Kab. Badung"
    },
    {
      "value": "L2",
      "label": "Kab. Balangan"
    },
    {
      "value": "L2",
      "label": "Kab. Bandung"
    },
    {
      "value": "L2",
      "label": "Kab. Bandung Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Banggai"
    },
    {
      "value": "L4",
      "label": "Kab. Banggai Kepulauan"
    },
    {
      "value": "L4",
      "label": "Kab. Banggai Laut"
    },
    {
      "value": "L2",
      "label": "Kab. Bangka"
    },
    {
      "value": "L2",
      "label": "Kab. Bangka Barat"
    },
    {
      "value": "L2",
      "label": "Kab. Bangka Selatan"
    },
    {
      "value": "L2",
      "label": "Kab. Bangka Tengah"
    },
    {
      "value": "L1",
      "label": "Kab. Bangkalan"
    },
    {
      "value": "L1",
      "label": "Kab. Bangli"
    },
    {
      "value": "L3",
      "label": "Kab. Banjar"
    },
    {
      "value": "L4",
      "label": "Kab. Banjarnegara"
    },
    {
      "value": "L4",
      "label": "Kab. Bantaeng"
    },
    {
      "value": "L4",
      "label": "Kab. Bantul"
    },
    {
      "value": "L3",
      "label": "Kab. Banyuasin"
    },
    {
      "value": "L4",
      "label": "Kab. Banyumas"
    },
    {
      "value": "L4",
      "label": "Kab. Banyuwangi"
    },
    {
      "value": "L3",
      "label": "Kab. Barito Kuala"
    },
    {
      "value": "L4",
      "label": "Kab. Barito Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Barito Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Barito Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Barru"
    },
    {
      "value": "L4",
      "label": "Kab. Batang"
    },
    {
      "value": "L2",
      "label": "Kab. Batanghari"
    },
    {
      "value": "L2",
      "label": "Kab. Batu Bara"
    },
    {
      "value": "L3",
      "label": "Kab. Bekasi"
    },
    {
      "value": "L3",
      "label": "Kab. Belitung"
    },
    {
      "value": "L3",
      "label": "Kab. Belitung Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Belu"
    },
    {
      "value": "L4",
      "label": "Kab. Bener Meriah"
    },
    {
      "value": "L2",
      "label": "Kab. Bengkalis"
    },
    {
      "value": "L4",
      "label": "Kab. Bengkayang"
    },
    {
      "value": "L4",
      "label": "Kab. Bengkulu Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Bengkulu Tengah"
    },
    {
      "value": "L3",
      "label": "Kab. Bengkulu Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Berau"
    },
    {
      "value": "L1",
      "label": "Kab. Biak Numfor"
    },
    {
      "value": "L4",
      "label": "Kab. Bima"
    },
    {
      "value": "L3",
      "label": "Kab. Bintan"
    },
    {
      "value": "L3",
      "label": "Kab. Bireuen"
    },
    {
      "value": "L4",
      "label": "Kab. Blitar"
    },
    {
      "value": "L4",
      "label": "Kab. Blora"
    },
    {
      "value": "L4",
      "label": "Kab. Boalemo"
    },
    {
      "value": "L3",
      "label": "Kab. Bogor"
    },
    {
      "value": "L4",
      "label": "Kab. Bojonegoro"
    },
    {
      "value": "L4",
      "label": "Kab. Bolaang Mongondow"
    },
    {
      "value": "L4",
      "label": "Kab. Bolaang Mongondow Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Bolaang Mongondow Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Bolaang Mongondow Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Bombana"
    },
    {
      "value": "L4",
      "label": "Kab. Bondowoso"
    },
    {
      "value": "L4",
      "label": "Kab. Bone"
    },
    {
      "value": "L4",
      "label": "Kab. Bone Bolango"
    },
    {
      "value": "L1",
      "label": "Kab. Boven Digoel"
    },
    {
      "value": "L4",
      "label": "Kab. Boyolali"
    },
    {
      "value": "L2",
      "label": "Kab. Brebes"
    },
    {
      "value": "L1",
      "label": "Kab. Buleleng"
    },
    {
      "value": "L4",
      "label": "Kab. Bulukumba"
    },
    {
      "value": "L4",
      "label": "Kab. Bulungan"
    },
    {
      "value": "L4",
      "label": "Kab. Bungo"
    },
    {
      "value": "L4",
      "label": "Kab. Buol"
    },
    {
      "value": "L1",
      "label": "Kab. Buru"
    },
    {
      "value": "L1",
      "label": "Kab. Buru Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Buton"
    },
    {
      "value": "L3",
      "label": "Kab. Buton Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Buton Tengah"
    },
    {
      "value": "L2",
      "label": "Kab. Buton Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Ciamis"
    },
    {
      "value": "L4",
      "label": "Kab. Cianjur"
    },
    {
      "value": "L3",
      "label": "Kab. Cilacap"
    },
    {
      "value": "L2",
      "label": "Kab. Cirebon"
    },
    {
      "value": "L2",
      "label": "Kab. Dairi"
    },
    {
      "value": "L1",
      "label": "Kab. Deiyai"
    },
    {
      "value": "L2",
      "label": "Kab. Deli Serdang"
    },
    {
      "value": "L4",
      "label": "Kab. Demak"
    },
    {
      "value": "L4",
      "label": "Kab. Dharmasraya"
    },
    {
      "value": "L1",
      "label": "Kab. Dogiyai"
    },
    {
      "value": "L3",
      "label": "Kab. Dompu"
    },
    {
      "value": "L4",
      "label": "Kab. Donggala"
    },
    {
      "value": "L3",
      "label": "Kab. Empat Lawang"
    },
    {
      "value": "L4",
      "label": "Kab. Ende"
    },
    {
      "value": "L4",
      "label": "Kab. Enrekang"
    },
    {
      "value": "L1",
      "label": "Kab. Fak Fak"
    },
    {
      "value": "L4",
      "label": "Kab. Flores Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Garut"
    },
    {
      "value": "L3",
      "label": "Kab. Gayo Lues"
    },
    {
      "value": "L1",
      "label": "Kab. Gianyar"
    },
    {
      "value": "L4",
      "label": "Kab. Gorontalo"
    },
    {
      "value": "L4",
      "label": "Kab. Gorontalo Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Gowa"
    },
    {
      "value": "L4",
      "label": "Kab. Gresik"
    },
    {
      "value": "L4",
      "label": "Kab. Grobogan"
    },
    {
      "value": "L4",
      "label": "Kab. Gunung Mas"
    },
    {
      "value": "L4",
      "label": "Kab. Gunungkidul"
    },
    {
      "value": "L1",
      "label": "Kab. Halmahera Barat"
    },
    {
      "value": "L1",
      "label": "Kab. Halmahera Selatan"
    },
    {
      "value": "L1",
      "label": "Kab. Halmahera Tengah"
    },
    {
      "value": "L1",
      "label": "Kab. Halmahera Timur"
    },
    { "value": "L1", "label": "Kab. Halmahera Utara" },
    {
      "value": "L2",
      "label": "Kab. Hulu Sungai Selatan"
    },
    {
      "value": "L2",
      "label": "Kab. Hulu Sungai Tengah"
    },
    {
      "value": "L3",
      "label": "Kab. Hulu Sungai Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Humbang Hasundutan"
    },
    {
      "value": "L4",
      "label": "Kab. Indragiri Hilir"
    },
    {
      "value": "L2",
      "label": "Kab. Indragiri Hulu"
    },
    {
      "value": "L2",
      "label": "Kab. Indramayu"
    },
    {
      "value": "L1",
      "label": "Kab. Intan Jaya"
    },
    {
      "value": "L1",
      "label": "Kab. Jayapura"
    },
    {
      "value": "L1",
      "label": "Kab. Jayawijaya"
    },
    {
      "value": "L4",
      "label": "Kab. Jember"
    },
    {
      "value": "L1",
      "label": "Kab. Jembrana"
    },
    {
      "value": "L4",
      "label": "Kab. Jeneponto"
    },
    {
      "value": "L4",
      "label": "Kab. Jepara"
    },
    {
      "value": "L4",
      "label": "Kab. Jombang"
    },
    {
      "value": "L1",
      "label": "Kab. Kaimana"
    },
    {
      "value": "L2",
      "label": "Kab. Kampar"
    },
    {
      "value": "L3",
      "label": "Kab. Kapuas"
    },
    {
      "value": "L4",
      "label": "Kab. Kapuas Hulu"
    },
    {
      "value": "L4",
      "label": "Kab. Karanganyar"
    },
    {
      "value": "L2",
      "label": "Kab. Karangasem"
    },
    {
      "value": "L4",
      "label": "Kab. Karawang"
    },
    {
      "value": "L3",
      "label": "Kab. Karimun"
    },
    {
      "value": "L2",
      "label": "Kab. Karo"
    },
    { "value": "L4", "label": "Kab. Katingan" },
    {
      "value": "L4",
      "label": "Kab. Kaur"
    },
    {
      "value": "L4",
      "label": "Kab. Kayong Utara"
    },
    {
      "value": "L2",
      "label": "Kab. Kebumen"
    },
    {
      "value": "L4",
      "label": "Kab. Kediri"
    },
    {
      "value": "L1",
      "label": "Kab. Keerom"
    },
    {
      "value": "L4",
      "label": "Kab. Kendal"
    },
    {
      "value": "L3",
      "label": "Kab. Kepahiang"
    },
    {
      "value": "L4",
      "label": "Kab. Kepulauan Anambas"
    },
    {
      "value": "L1",
      "label": "Kab. Kepulauan Aru"
    },
    {
      "value": "L2",
      "label": "Kab. Kepulauan Mentawai"
    },
    {
      "value": "L3",
      "label": "Kab. Kepulauan Meranti"
    },
    {
      "value": "L4",
      "label": "Kab. Kepulauan Sangihe"
    },
    {
      "value": "L4",
      "label": "Kab. Kepulauan Selayar"
    },
    {
      "value": "L3",
      "label": "Kab. Kepulauan Seribu"
    },
    {
      "value": "L1",
      "label": "Kab. Kepulauan Sula"
    },
    {
      "value": "L4",
      "label": "Kab. Kepulauan Talaud"
    },
    {
      "value": "L1",
      "label": "Kab. Kepulauan Yapen"
    },
    {
      "value": "L4",
      "label": "Kab. Kerinci"
    },
    {
      "value": "L4",
      "label": "Kab. Ketapang"
    },
    {
      "value": "L4",
      "label": "Kab. Klaten"
    },
    {
      "value": "L1",
      "label": "Kab. Klungkung"
    },
    {
      "value": "L4",
      "label": "Kab. Kolaka"
    },
    { "value": "L4", "label": "Kab. Kolaka Timur" },
    {
      "value": "L4",
      "label": "Kab. Kolaka Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Konawe"
    },
    {
      "value": "L3",
      "label": "Kab. Konawe Kepulauan"
    },
    {
      "value": "L4",
      "label": "Kab. Konawe Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Konawe Utara"
    },
    {
      "value": "L2",
      "label": "Kab. Kotabaru"
    },
    {
      "value": "L4",
      "label": "Kab. Kotawaringin Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Kotawaringin Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Kuantan Singingi"
    },
    {
      "value": "L4",
      "label": "Kab. Kubu Raya"
    },
    {
      "value": "L4",
      "label": "Kab. Kudus"
    },
    {
      "value": "L4",
      "label": "Kab. Kulon Progo"
    },
    {
      "value": "L2",
      "label": "Kab. Kuningan"
    },
    {
      "value": "L4",
      "label": "Kab. Kupang"
    },
    {
      "value": "L4",
      "label": "Kab. Kutai Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Kutai Kartanegara"
    },
    {
      "value": "L4",
      "label": "Kab. Kutai Timur"
    },
    {
      "value": "L3",
      "label": "Kab. Labuhanbatu"
    },
    {
      "value": "L4",
      "label": "Kab. Labuhanbatu Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Labuhanbatu Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Lahat"
    },
    {
      "value": "L4",
      "label": "Kab. Lamandau"
    },
    {
      "value": "L4",
      "label": "Kab. Lamongan"
    },
    {
      "value": "L4",
      "label": "Kab. Lampung Barat"
    },
    {
      "value": "L2",
      "label": "Kab. Lampung Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Lampung Tengah"
    },
    {
      "value": "L4",
      "label": "Kab. Lampung Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Lampung Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Landak"
    },
    {
      "value": "L2",
      "label": "Kab. Langkat"
    },
    {
      "value": "L1",
      "label": "Kab. Lanny Jaya"
    },
    {
      "value": "L2",
      "label": "Kab. Lebak"
    },
    {
      "value": "L4",
      "label": "Kab. Lebong"
    },
    {
      "value": "L4",
      "label": "Kab. Lembata"
    },
    {
      "value": "L2",
      "label": "Kab. Lima Puluh Kota"
    },
    {
      "value": "L4",
      "label": "Kab. Lingga"
    },
    {
      "value": "L1",
      "label": "Kab. Lombok Barat"
    },
    {
      "value": "L1",
      "label": "Kab. Lombok Tengah"
    },
    {
      "value": "L1",
      "label": "Kab. Lombok Timur"
    },
    {
      "value": "L3",
      "label": "Kab. Lombok Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Lumajang"
    },
    {
      "value": "L4",
      "label": "Kab. Luwu"
    },
    {
      "value": "L3",
      "label": "Kab. Luwu Timur"
    },
    {
      "value": "L4",
      "label": "Kab. Luwu Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Madiun"
    },
    {
      "value": "L4",
      "label": "Kab. Magelang"
    },
    {
      "value": "L4",
      "label": "Kab. Magetan"
    },
    {
      "value": "L4",
      "label": "Kab. Mahakam Ulu"
    },
    {
      "value": "L2",
      "label": "Kab. Majalengka"
    },
    {
      "value": "L3",
      "label": "Kab. Majene"
    },
    { "value": "L4", "label": "Kab. Malaka" },
    {
      "value": "L4",
      "label": "Kab. Malang"
    },
    {
      "value": "L4",
      "label": "Kab. Malinau"
    },
    {
      "value": "L1",
      "label": "Kab. Maluku Barat Daya"
    },
    {
      "value": "L1",
      "label": "Kab. Maluku Tengah"
    },
    {
      "value": "L1",
      "label": "Kab. Maluku Tenggara"
    },
    {
      "value": "L1",
      "label": "Kab. Maluku Tenggara Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Mamasa"
    },
    {
      "value": "L1",
      "label": "Kab. Mamberamo Raya"
    },
    {
      "value": "L1",
      "label": "Kab. Mamberamo Tengah"
    },
    {
      "value": "L4",
      "label": "Kab. Mamuju"
    },
    {
      "value": "L3",
      "label": "Kab. Mamuju Tengah"
    },
    {
      "value": "L4",
      "label": "Kab. Mamuju Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Mandailing Natal"
    },
    {
      "value": "L4",
      "label": "Kab. Manggarai"
    },
    {
      "value": "L4",
      "label": "Kab. Manggarai Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Manggarai Timur"
    },
    {
      "value": "L1",
      "label": "Kab. Manokwari"
    },
    {
      "value": "L1",
      "label": "Kab. Manokwari Selatan"
    },
    {
      "value": "L1",
      "label": "Kab. Mappi"
    },
    {
      "value": "L4",
      "label": "Kab. Maros"
    },
    {
      "value": "L1",
      "label": "Kab. Maybrat"
    },
    {
      "value": "L4",
      "label": "Kab. Melawi"
    },
    {
      "value": "L4",
      "label": "Kab. Mempawah"
    },
    {
      "value": "L4",
      "label": "Kab. Merangin"
    },
    {
      "value": "L1",
      "label": "Kab. Merauke"
    },
    {
      "value": "L4",
      "label": "Kab. Mesuji"
    },
    {
      "value": "L1",
      "label": "Kab. Mimika"
    },
    {
      "value": "L4",
      "label": "Kab. Minahasa"
    },
    {
      "value": "L4",
      "label": "Kab. Minahasa Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Minahasa Tenggara"
    },
    {
      "value": "L4",
      "label": "Kab. Minahasa Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Mojokerto"
    },
    {
      "value": "L4",
      "label": "Kab. Morowali"
    },
    {
      "value": "L4",
      "label": "Kab. Morowali Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Muara Enim"
    },
    {
      "value": "L2",
      "label": "Kab. Muaro Jambi"
    },
    {
      "value": "L4",
      "label": "Kab. Muko Muko"
    },
    {
      "value": "L3",
      "label": "Kab. Muna"
    },
    {
      "value": "L3",
      "label": "Kab. Muna Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Murung Raya"
    },
    {
      "value": "L2",
      "label": "Kab. Musi Banyuasin"
    },
    {
      "value": "L3",
      "label": "Kab. Musi Rawas"
    },
    {
      "value": "L3",
      "label": "Kab. Musi Rawas Utara"
    },
    {
      "value": "L1",
      "label": "Kab. Nabire"
    },
    {
      "value": "L3",
      "label": "Kab. Nagan Raya"
    },
    {
      "value": "L4",
      "label": "Kab. Nagekeo"
    },
    {
      "value": "L4",
      "label": "Kab. Natuna"
    },
    {
      "value": "L1",
      "label": "Kab. Nduga"
    },
    {
      "value": "L4",
      "label": "Kab. Ngada"
    },
    {
      "value": "L4",
      "label": "Kab. Nganjuk"
    },
    {
      "value": "L4",
      "label": "Kab. Ngawi"
    },
    {
      "value": "L4",
      "label": "Kab. Nias"
    },
    {
      "value": "L3",
      "label": "Kab. Nias Barat"
    },
    {
      "value": "L3",
      "label": "Kab. Nias Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Nias Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Nunukan"
    },
    {
      "value": "L2",
      "label": "Kab. Ogan Ilir"
    },
    {
      "value": "L2",
      "label": "Kab. Ogan Komering Ilir"
    },
    {
      "value": "L4",
      "label": "Kab. Ogan Komering Ulu"
    },
    {
      "value": "L4",
      "label": "Kab. Ogan Komering Ulu Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Ogan Komering Ulu Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Pacitan"
    },
    {
      "value": "L4",
      "label": "Kab. Padang Lawas"
    },
    {
      "value": "L3",
      "label": "Kab. Padang Lawas Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Padang Pariaman"
    },
    {
      "value": "L4",
      "label": "Kab. Pahuwato"
    },
    {
      "value": "L4",
      "label": "Kab. Pakpak Bharat"
    },
    {
      "value": "L1",
      "label": "Kab. Pamekasan"
    },
    {
      "value": "L2",
      "label": "Kab. Pandeglang"
    },
    {
      "value": "L3",
      "label": "Kab. Pangandaran"
    },
    {
      "value": "L4",
      "label": "Kab. Pangkajene Kepulauan"
    },
    {
      "value": "L1",
      "label": "Kab. Paniai"
    },
    {
      "value": "L4",
      "label": "Kab. Parigi Moutong"
    },
    {
      "value": "L3",
      "label": "Kab. Pasaman"
    },
    {
      "value": "L2",
      "label": "Kab. Pasaman Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Paser"
    },
    {
      "value": "L4",
      "label": "Kab. Pasuruan"
    },
    {
      "value": "L4",
      "label": "Kab. Pati"
    },
    {
      "value": "L1",
      "label": "Kab. Pegunungan Arfak"
    },
    {
      "value": "L1",
      "label": "Kab. Pegunungan Bintang"
    },
    {
      "value": "L4",
      "label": "Kab. Pekalongan"
    },
    {
      "value": "L2",
      "label": "Kab. Pelalawan"
    },
    {
      "value": "L2",
      "label": "Kab. Pemalang"
    },
    {
      "value": "L4",
      "label": "Kab. Penajam Paser Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Penukal Abab Lematang Ilir"
    },
    {
      "value": "L2",
      "label": "Kab. Pesawaran"
    },
    {
      "value": "L4",
      "label": "Kab. Pesisir Barat"
    },
    {
      "value": "L3",
      "label": "Kab. Pesisir Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Pidie"
    },
    {
      "value": "L3",
      "label": "Kab. Pidie Jaya"
    },
    {
      "value": "L3",
      "label": "Kab. Pinrang"
    },
    {
      "value": "L3",
      "label": "Kab. Polewali Mandar"
    },
    {
      "value": "L4",
      "label": "Kab. Ponorogo"
    },
    {
      "value": "L4",
      "label": "Kab. Poso"
    },
    {
      "value": "L3",
      "label": "Kab. Pringsewu"
    },
    {
      "value": "L3",
      "label": "Kab. Probolinggo"
    },
    {
      "value": "L2",
      "label": "Kab. Pulang Pisau"
    },
    {
      "value": "L1",
      "label": "Kab. Pulau Morotai"
    },
    { "value": "L1", "label": "Kab. Pulau Taliabu" },
    {
      "value": "L1",
      "label": "Kab. Puncak"
    },
    {
      "value": "L1",
      "label": "Kab. Puncak Jaya"
    },
    {
      "value": "L4",
      "label": "Kab. Purbalingga"
    },
    {
      "value": "L2",
      "label": "Kab. Purwakarta"
    },
    {
      "value": "L4",
      "label": "Kab. Purworejo"
    },
    {
      "value": "L1",
      "label": "Kab. Raja Ampat"
    },
    {
      "value": "L4",
      "label": "Kab. Rejang Lebong"
    },
    {
      "value": "L4",
      "label": "Kab. Rembang"
    },
    {
      "value": "L2",
      "label": "Kab. Rokan Hilir"
    },
    {
      "value": "L4",
      "label": "Kab. Rokan Hulu"
    },
    {
      "value": "L4",
      "label": "Kab. Rote Ndao"
    },
    {
      "value": "L4",
      "label": "Kab. Sabu Raijua"
    },
    {
      "value": "L4",
      "label": "Kab. Sambas"
    },
    {
      "value": "L3",
      "label": "Kab. Samosir"
    },
    {
      "value": "L1",
      "label": "Kab. Sampang"
    },
    {
      "value": "L4",
      "label": "Kab. Sanggau"
    },
    {
      "value": "L1",
      "label": "Kab. Sarmi"
    },
    {
      "value": "L3",
      "label": "Kab. Sarolangun"
    },
    {
      "value": "L4",
      "label": "Kab. Sekadau"
    },
    {
      "value": "L3",
      "label": "Kab. Seluma"
    },
    { "value": "L4", "label": "Kab. Semarang" },
    {
      "value": "L1",
      "label": "Kab. Seram Bagian Barat"
    },
    {
      "value": "L1",
      "label": "Kab. Seram Bagian Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Serang"
    },
    {
      "value": "L2",
      "label": "Kab. Serdang Bedagai"
    },
    {
      "value": "L4",
      "label": "Kab. Seruyan"
    },
    {
      "value": "L2",
      "label": "Kab. Siak"
    },
    {
      "value": "L4",
      "label": "Kab. Siau Tagulandang Biaro"
    },
    {
      "value": "L3",
      "label": "Kab. Sidenreng Rappang"
    },
    {
      "value": "L3",
      "label": "Kab. Sidoarjo"
    },
    {
      "value": "L4",
      "label": "Kab. Sigi"
    },
    {
      "value": "L2",
      "label": "Kab. Sijunjung"
    },
    {
      "value": "L4",
      "label": "Kab. Sikka"
    },
    {
      "value": "L3",
      "label": "Kab. Simalungun"
    },
    {
      "value": "L4",
      "label": "Kab. Simeulue"
    },
    { "value": "L3", "label": "Kab. Sinjai" },
    {
      "value": "L4",
      "label": "Kab. Sintang"
    },
    {
      "value": "L4",
      "label": "Kab. Situbondo"
    },
    {
      "value": "L4",
      "label": "Kab. Sleman"
    },
    {
      "value": "L2",
      "label": "Kab. Solok"
    },
    {
      "value": "L2",
      "label": "Kab. Solok Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Soppeng"
    },
    {
      "value": "L1",
      "label": "Kab. Sorong"
    },
    {
      "value": "L1",
      "label": "Kab. Sorong Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Sragen"
    },
    {
      "value": "L2",
      "label": "Kab. Subang"
    },
    {
      "value": "L4",
      "label": "Kab. Sukabumi"
    },
    {
      "value": "L4",
      "label": "Kab. Sukamara"
    },
    {
      "value": "L4",
      "label": "Kab. Sukoharjo"
    },
    {
      "value": "L4",
      "label": "Kab. Sumba Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Sumba Barat Daya"
    },
    {
      "value": "L4",
      "label": "Kab. Sumba Tengah"
    },
    {
      "value": "L4",
      "label": "Kab. Sumba Timur"
    },
    {
      "value": "L2",
      "label": "Kab. Sumbawa"
    },
    {
      "value": "L3",
      "label": "Kab. Sumbawa Barat"
    },
    {
      "value": "L3",
      "label": "Kab. Sumedang"
    },
    {
      "value": "L1",
      "label": "Kab. Sumenep"
    },
    {
      "value": "L1",
      "label": "Kab. Supiori"
    },
    {
      "value": "L2",
      "label": "Kab. Tabalong"
    },
    {
      "value": "L1",
      "label": "Kab. Tabanan"
    },
    {
      "value": "L4",
      "label": "Kab. Takalar"
    },
    {
      "value": "L1",
      "label": "Kab. Tambrauw"
    },
    {
      "value": "L4",
      "label": "Kab. Tana Tidung"
    },
    {
      "value": "L3",
      "label": "Kab. Tana Toraja"
    },
    {
      "value": "L2",
      "label": "Kab. Tanah Bumbu"
    },
    {
      "value": "L3",
      "label": "Kab. Tanah Datar"
    },
    {
      "value": "L2",
      "label": "Kab. Tanah Laut"
    },
    {
      "value": "L2",
      "label": "Kab. Tangerang"
    },
    {
      "value": "L4",
      "label": "Kab. Tanggamus"
    },
    {
      "value": "L2",
      "label": "Kab. Tanjung Jabung Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Tanjung Jabung Timur"
    },
    {
      "value": "L3",
      "label": "Kab. Tapanuli Selatan"
    },
    {
      "value": "L3",
      "label": "Kab. Tapanuli Tengah"
    },
    {
      "value": "L3",
      "label": "Kab. Tapanuli Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Tapin"
    },
    {
      "value": "L3",
      "label": "Kab. Tasikmalaya"
    },
    {
      "value": "L3",
      "label": "Kab. Tebo"
    },
    {
      "value": "L2",
      "label": "Kab. Tegal"
    },
    {
      "value": "L1",
      "label": "Kab. Teluk Bintuni"
    },
    {
      "value": "L1",
      "label": "Kab. Teluk Wondama"
    },
    {
      "value": "L4",
      "label": "Kab. Temanggung"
    },
    {
      "value": "L4",
      "label": "Kab. Timor Tengah Selatan"
    },
    {
      "value": "L4",
      "label": "Kab. Timor Tengah Utara"
    },
    {
      "value": "L3",
      "label": "Kab. Toba Samosir"
    },
    {
      "value": "L4",
      "label": "Kab. Tojo Una Una"
    },
    {
      "value": "L4",
      "label": "Kab. Toli Toli"
    },
    {
      "value": "L1",
      "label": "Kab. Tolikara"
    },
    {
      "value": "L4",
      "label": "Kab. Toraja Utara"
    },
    {
      "value": "L4",
      "label": "Kab. Trenggalek"
    },
    {
      "value": "L4",
      "label": "Kab. Tuban"
    },
    {
      "value": "L4",
      "label": "Kab. Tulang Bawang"
    },
    {
      "value": "L4",
      "label": "Kab. Tulang Bawang Barat"
    },
    {
      "value": "L4",
      "label": "Kab. Tulungagung"
    },
    {
      "value": "L3",
      "label": "Kab. Wajo"
    },
    {
      "value": "L2",
      "label": "Kab. Wakatobi"
    },
    {
      "value": "L1",
      "label": "Kab. Waropen"
    },
    {
      "value": "L4",
      "label": "Kab. Way Kanan"
    },
    {
      "value": "L4",
      "label": "Kab. Wonogiri"
    },
    {
      "value": "L4",
      "label": "Kab. Wonosobo"
    },
    {
      "value": "L1",
      "label": "Kab. Yahukimo"
    },
    {
      "value": "L1",
      "label": "Kab. Yalimo"
    },
    {
      "value": "L1",
      "label": "Kota Ambon"
    },
    {
      "value": "L4",
      "label": "Kota Balikpapan"
    },
    {
      "value": "L3",
      "label": "Kota Banda Aceh"
    },
    {
      "value": "L4",
      "label": "Kota Bandar Lampung"
    },
    {
      "value": "L2",
      "label": "Kota Bandung"
    },
    { "value": "L3", "label": "Kota Banjar" },
    {
      "value": "L3",
      "label": "Kota Banjarbaru"
    },
    {
      "value": "L3",
      "label": "Kota Banjarmasin"
    },
    {
      "value": "L3",
      "label": "Kota Batam"
    },
    {
      "value": "L4",
      "label": "Kota Batu"
    },
    {
      "value": "L3",
      "label": "Kota Bau Bau"
    },
    {
      "value": "L3",
      "label": "Kota Bekasi"
    },
    {
      "value": "L3",
      "label": "Kota Bengkulu"
    },
    {
      "value": "L3",
      "label": "Kota Bima"
    },
    {
      "value": "L2",
      "label": "Kota Binjai"
    },
    {
      "value": "L4",
      "label": "Kota Bitung"
    },
    {
      "value": "L4",
      "label": "Kota Blitar"
    },
    {
      "value": "L3",
      "label": "Kota Bogor"
    },
    {
      "value": "L4",
      "label": "Kota Bontang"
    },
    {
      "value": "L3",
      "label": "Kota Bukittinggi"
    },
    {
      "value": "L2",
      "label": "Kota Cilegon"
    },
    {
      "value": "L3",
      "label": "Kota Cimahi"
    },
    {
      "value": "L2",
      "label": "Kota Cirebon"
    },
    {
      "value": "L2",
      "label": "Kota Denpasar"
    },
    {
      "value": "L3",
      "label": "Kota Depok"
    },
    {
      "value": "L3",
      "label": "Kota Dumai"
    },
    {
      "value": "L4",
      "label": "Kota Gorontalo"
    },
    {
      "value": "L3",
      "label": "Kota Gunungsitoli"
    },
    { "value": "L2", "label": "Kota Jakarta Barat" },
    {
      "value": "L2",
      "label": "Kota Jakarta Pusat"
    },
    {
      "value": "L2",
      "label": "Kota Jakarta Selatan"
    },
    {
      "value": "L2",
      "label": "Kota Jakarta Timur"
    },
    {
      "value": "L2",
      "label": "Kota Jakarta Utara"
    },
    {
      "value": "L3",
      "label": "Kota Jambi"
    },
    {
      "value": "L1",
      "label": "Kota Jayapura"
    },
    {
      "value": "L4",
      "label": "Kota Kediri"
    },
    {
      "value": "L4",
      "label": "Kota Kendari"
    },
    {
      "value": "L4",
      "label": "Kota Kotamobagu"
    },
    {
      "value": "L4",
      "label": "Kota Kupang"
    },
    {
      "value": "L4",
      "label": "Kota Langsa"
    },
    {
      "value": "L3",
      "label": "Kota Lhokseumawe"
    },
    {
      "value": "L3",
      "label": "Kota Lubuk Linggau"
    },
    {
      "value": "L4",
      "label": "Kota Madiun"
    },
    {
      "value": "L4",
      "label": "Kota Magelang"
    },
    {
      "value": "L4",
      "label": "Kota Makassar"
    },
    {
      "value": "L4",
      "label": "Kota Malang"
    },
    {
      "value": "L4",
      "label": "Kota Manado"
    },
    {
      "value": "L1",
      "label": "Kota Mataram"
    },
    {
      "value": "L2",
      "label": "Kota Medan"
    },
    {
      "value": "L3",
      "label": "Kota Metro"
    },
    {
      "value": "L4",
      "label": "Kota Mojokerto"
    },
    {
      "value": "L2",
      "label": "Kota Padang"
    },
    {
      "value": "L3",
      "label": "Kota Padang Panjang"
    },
    {
      "value": "L3",
      "label": "Kota Padangsidimpuan"
    },
    {
      "value": "L3",
      "label": "Kota Pagar Alam"
    },
    { "value": "L3", "label": "Kota Palangkaraya" },
    {
      "value": "L3",
      "label": "Kota Palembang"
    },
    {
      "value": "L4",
      "label": "Kota Palopo"
    },
    {
      "value": "L4",
      "label": "Kota Palu"
    },
    {
      "value": "L2",
      "label": "Kota Pangkal Pinang"
    },
    {
      "value": "L3",
      "label": "Kota Pare Pare"
    },
    {
      "value": "L2",
      "label": "Kota Pariaman"
    },
    {
      "value": "L4",
      "label": "Kota Pasuruan"
    },
    {
      "value": "L2",
      "label": "Kota Payakumbuh"
    },
    {
      "value": "L4",
      "label": "Kota Pekalongan"
    },
    {
      "value": "L3",
      "label": "Kota Pekanbaru"
    },
    {
      "value": "L3",
      "label": "Kota Pematangsiantar"
    },
    {
      "value": "L4",
      "label": "Kota Pontianak"
    },
    {
      "value": "L3",
      "label": "Kota Prabumulih"
    },
    {
      "value": "L3",
      "label": "Kota Probolinggo"
    },
    {
      "value": "L2",
      "label": "Kota Sabang"
    },
    {
      "value": "L4",
      "label": "Kota Salatiga"
    },
    {
      "value": "L4",
      "label": "Kota Samarinda"
    },
    {
      "value": "L4",
      "label": "Kota Sawahlunto"
    },
    {
      "value": "L4",
      "label": "Kota Semarang"
    },
    {
      "value": "L2",
      "label": "Kota Serang"
    },
    {
      "value": "L3",
      "label": "Kota Sibolga"
    },
    {
      "value": "L4",
      "label": "Kota Singkawang"
    },
    { "value": "L2", "label": "Kota Solok" },
    {
      "value": "L1",
      "label": "Kota Sorong"
    },
    {
      "value": "L4",
      "label": "Kota Subulussalam"
    },
    {
      "value": "L4",
      "label": "Kota Sukabumi"
    },
    {
      "value": "L4",
      "label": "Kota Sungai Penuh"
    },
    {
      "value": "L3",
      "label": "Kota Surabaya"
    },
    {
      "value": "L4",
      "label": "Kota Surakarta"
    },
    {
      "value": "L2",
      "label": "Kota Tangerang"
    },
    {
      "value": "L2",
      "label": "Kota Tangerang Selatan"
    },
    {
      "value": "L2",
      "label": "Kota Tanjung Balai"
    },
    {
      "value": "L2",
      "label": "Kota Tanjung Pinang"
    },
    {
      "value": "L4",
      "label": "Kota Tarakan"
    },
    {
      "value": "L3",
      "label": "Kota Tasikmalaya"
    },
    {
      "value": "L2",
      "label": "Kota Tebing Tinggi"
    },
    {
      "value": "L2",
      "label": "Kota Tegal"
    },
    {
      "value": "L1",
      "label": "Kota Ternate"
    },
    {
      "value": "L1",
      "label": "Kota Tidore Kepulauan"
    },
    {
      "value": "L4",
      "label": "Kota Tomohon"
    },
    {
      "value": "L1",
      "label": "Kota Tual"
    },
    {
      "value": "L4",
      "label": "Kota Yogyakarta"
    }
  ]
}





# cek stock kuota akrab
curl 'https://ics-store.my.id/api.php?action=fetchProducts&type=bpa' \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.8' \
  -H 'priority: u=1, i' \
  -H 'referer: https://ics-store.my.id/' \
  -H 'sec-ch-ua: "Chromium";v="142", "Brave";v="142", "Not_A Brand";v="99"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Linux"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'sec-gpc: 1' \
  -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36'

{
    "success": true,
    "type": "bpa",
    "title": "Daftar Produk Harian",
    "providers_used": [
        "FMAX",
        "KUBER",
        "KUBERV2"
    ],
    "data": {
        "ready": [],
        "empty": []
    }
}


curl 'https://ics-store.my.id/api.php?action=fetchProducts&type=xda' \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.8' \
  -H 'priority: u=1, i' \
  -H 'referer: https://ics-store.my.id/' \
  -H 'sec-ch-ua: "Chromium";v="142", "Brave";v="142", "Not_A Brand";v="99"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Linux"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'sec-gpc: 1' \
  -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36'

{
    "success": true,
    "type": "xda",
    "title": "Daftar Produk BulananV2",
    "providers_used": [
        "FMAX",
        "KUBER",
        "KUBERV2"
    ],
    "data": {
        "ready": [],
        "empty": []
    }
}

curl 'https://ics-store.my.id/api.php?action=fetchProducts&type=xla' \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.8' \
  -H 'priority: u=1, i' \
  -H 'referer: https://ics-store.my.id/' \
  -H 'sec-ch-ua: "Chromium";v="142", "Brave";v="142", "Not_A Brand";v="99"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Linux"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'sec-gpc: 1' \
  -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36'

{
    "success": true,
    "type": "xla",
    "title": "Daftar Produk Bulanan",
    "providers_used": [
        "FMAX",
        "KUBER",
        "KUBERV2"
    ],
    "data": {
        "ready": [],
        "empty": [
            {
                "kode_produk": "XLA14",
                "nama_produk": "SuperMini",
                "deskripsi": "Details Kuota :\r\nAREA 1 = 13 - 15 GB\r\nAREA 2 = 15 - 17 GB\r\nAREA 3 = 20 - 22 GB\r\nAREA 4 = 30 - 32 GB\r\n\r\nnoted :\r\n~ rewards tidak masuk, tunngu 1 x 24 jam, baru komplen\r\n~ official, resmi, bergaransi\r\n",
                "harga_original": 41000,
                "harga_final": 42000,
                "stok": 0,
                "type": "xla"
            },
            {
                "kode_produk": "XLA32",
                "nama_produk": "Mini",
                "deskripsi": "Details Kuota :\r\n~ AREA 1 : 31.75 GB - 33.75 GB\r\n~ AREA 2 : 33.75 - 35.75 GB\r\n~ AREA 3 : 38.75 - 40.75 GB\r\n~ AREA 4 : 48.75 - 50.75 GB\r\n\r\nnoted :\r\n~ rewards tidak masuk, tunngu 1 x 24 jam, baru komplen\r\n~ official, resmi, bergaransi\r\n",
                "harga_original": 52000,
                "harga_final": 53000,
                "stok": 0,
                "type": "xla"
            },
            {
                "kode_produk": "XLA39",
                "nama_produk": "Big ",
                "deskripsi": "Details Kuota :\r\n~ AREA 1 : 38 GB - 40 GB\r\n~ AREA 2 : 40 GB - 42 GB\r\n~ AREA 3 : 45 GB - 47 GB\r\n~ AREA 4 : 55 GB - 57 GB\r\n\r\nnoted :\r\n~ rewards tidak masuk, tunngu 1 x 24 jam, baru komplen\r\n~ official, resmi, bergaransi\r\n",
                "harga_original": 57000,
                "harga_final": 58000,
                "stok": 0,
                "type": "xla"
            },
            {
                "kode_produk": "XLA51",
                "nama_produk": "Jumbo V2",
                "deskripsi": "Details Kuota :\r\nAREA 1 = 50.5 - 52.5 GB\r\nAREA 2 = 52.5 - 54.5 GB\r\nAREA 3 = 57.5 - 59.5 GB\r\nAREA 4 = 67.5 - 69.5 GB\r\nnoted :\r\n~ rewards tidak masuk, tunngu 1 x 24 jam, baru komplen\r\n~ official, resmi, bergaransi\r\n",
                "harga_original": 67000,
                "harga_final": 68000,
                "stok": 0,
                "type": "xla"
            },
            {
                "kode_produk": "XLA65",
                "nama_produk": "JUMBO",
                "deskripsi": "Details Kuota :\r\n~ Area 1 = 65 GB\r\n~ Area 2 = 70 GB\r\n~ Area 3 = 83 GB\r\n~ Area 4 = 123 GB\r\nnoted :\r\n~ rewards tidak masuk, tunngu 1 x 24 jam, baru komplen\r\n~ official, resmi, bergaransi\r\n",
                "harga_original": 81000,
                "harga_final": 82000,
                "stok": 0,
                "type": "xla"
            },
            {
                "kode_produk": "XLA89",
                "nama_produk": "MegaBig",
                "deskripsi": "Details Kuota :\r\n~ AREA 1 : 88 GB - 90 GB\r\n~ AREA 2 : 90 GB - 92 GB\r\n~ AREA 3 : 95 GB - 97 GB\r\n~ AREA 4 : 105 GB - 107 GB\r\nnoted :\r\n~ rewards tidak masuk, tunngu 1 x 24 jam, baru komplen\r\n~ official, resmi, bergaransi\r\n",
                "harga_original": 89000,
                "harga_final": 90000,
                "stok": 0,
                "type": "xla"
            }
        ]
    }
}
