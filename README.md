# TorScraper

TorScraper, Tor ağı üzerindeki .onion sitelerini otomatik olarak tarayan bir Cyber Threat Intelligence (CTI) aracıdır. Go (Golang) ile geliştirilmiştir.

## Özellikler

- ✅ Tor SOCKS5 proxy üzerinden anonim bağlantı
- ✅ Toplu hedef tarama (YAML dosyasından)
- ✅ **Otomatik screenshot (ekran görüntüsü) alma** 
- ✅ HTML içerik kaydetme
- ✅ Otomatik hata yönetimi
- ✅ Detaylı loglama sistemi
- ✅ Tor bağlantı doğrulama

## Gereksinimler

1. **Go 1.18 veya üzeri**
   ```bash
   go version
   ```

2. **Tor Browser veya Tor Service**
   - Windows: Tor Browser indirin ve çalıştırın
   - Linux: `sudo apt install tor && sudo systemctl start tor`
   - Tor SOCKS5 proxy'si 127.0.0.1:9050 veya 9150 portunda çalışmalıdır

3. **Google Chrome veya Chromium** (Screenshot için)
   - Windows: Chrome otomatik bulunur
   - Linux: `sudo apt install chromium-browser`

## Kurulum

1. Go 1.18+ yüklü olduğundan emin olun
2. Bağımlılıkları yükleyin:
   ```bash
   cd TorScraper
   go mod download



**Hızlı başlangıç**:
```bash
# 1. Tor Browser'ı başlat
# 2. Programı çalıştır
go run main.go
```

## Çıktılar

Program çalıştıktan sonra şu dosyalar oluşturulur:

- **output/** klasöründe:
  - `*.html` - İndirilen HTML dosyaları
  - `*.png` - Screenshot 
- **scan_report.log** - Tarama raporu 
