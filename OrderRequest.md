# Order Request Flow
Order Request Flow adalah proses dimana pelanggan dapat memesan barang-barang yang dijual oleh merchant. Berikut adalah flow dari Order Request:

1. User Membuat Order Request:
   - User memilih produk dan lokasi pengiriman.
   - Sistem mencari merchant terdekat yang menjual produk tersebut.
   - Sistem mengirim order request ke merchant terdekat.
   - ```order_request.CreateOrderRequest```
   - ```order_request.GetAvailableMerchant```
2. Merchant Mengambil Order:
   - Merchant melihat order request yang tersedia.
   - Merchant mengambil order dan menghitung harga + ongkir.
   - Merchant mengirim penawaran kembali ke sistem.
   - ```offer.CreateOffer```
3. Sistem Menampilkan Penawaran ke User:
   - Sistem menampilkan penawaran dari merchant yang telah mengambil order.
   - User memilih penawaran yang diinginkan.
   - ```offer.CreateOffer```
4. Batas Waktu Pengambilan Order:
   - Jika tidak ada merchant yang mengambil order dalam waktu tertentu, order request dibatalkan.
5. User Mengambil Order:
   - User memilih penawaran yang diinginkan.
   - Sistem mengirim konfirmasi ke merchant yang membuat penawaran.
   - Merchant mengirimkan produk ke alamat yang ditentukan.
6. User Menerima Barang:
   - User menerima barang yang dikirim.
   - User memberikan konfirmasi ke sistem bahwa barang telah diterima.
7. Sistem Membuat Data Penjualan:
   - Setelah user memberikan konfirmasi penerimaan barang, sistem otomatis membuat entri data penjualan.
   - Data penjualan mencakup informasi produk, jumlah, harga, dan tanggal transaksi.
   - Sistem mengupdate stok barang di gudang dan menyesuaikan laporan keuangan.
8. User Memberikan Feedback dan Reputasi:
   - Setelah user memberikan konfirmasi penerimaan barang, user diminta untuk memberikan feedback dan reputasi terhadap merchant.
   - Feedback dan reputasi digunakan untuk memberikan evaluasi terhadap merchant dan meningkatkan kualitas layanan.
9. Skenario Jika Barang Yang Dipesan Tidak Sesuai:
    - Jika user menerima barang yang tidak sesuai dengan pesanan, user dapat mengajukan keluhan ke sistem.
    - Sistem akan menghubungi merchant dan meminta merchant untuk mengirimkan ulang atau mengganti barang yang sesuai.
    - Jika merchant tidak mengirimkan ulang atau mengganti barang yang sesuai, maka sistem akan mengembalikan dana ke user.
