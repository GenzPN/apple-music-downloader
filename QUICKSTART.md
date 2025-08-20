# 🚀 Hướng dẫn nhanh - Apple Music Downloader

## ⚡ Bắt đầu trong 5 phút

### 1. Cài đặt dependencies

```bash
# Cài đặt Go dependencies
go mod download

# Cài đặt FFmpeg (Ubuntu/Debian)
sudo apt update && sudo apt install ffmpeg gpac

# Trên Windows: Tải FFmpeg từ https://ffmpeg.org/download.html
# Trên macOS: brew install ffmpeg gpac
```

### 2. Cấu hình

```bash
# Sao chép file cấu hình
cp config.yaml.example config.yaml

# Chỉnh sửa config.yaml với tokens của bạn
```

### 3. Lấy tokens (Quan trọng!)

#### Media User Token:
1. Mở [Apple Music](https://music.apple.com) và đăng nhập
2. Mở Developer Tools (F12)
3. Vào tab **Network**
4. Tìm request đến `amp-api.music.apple.com`
5. Copy giá trị của header `media-user-token`

#### Authorization Token:
- Thường tự động lấy được
- Nếu cần: Tìm header `authorization` trong request đến `amp-api.music.apple.com`

### 4. Chạy ứng dụng

#### Web Server (Khuyến nghị):
```bash
# Chạy server
go run server_main.go server.go main.go -port 8080

# Mở trình duyệt: http://localhost:8080
```

#### Command Line:
```bash
# Tải xuống album
go run main.go https://music.apple.com/us/album/album-name/id123456789

# Tải xuống với chất lượng cụ thể
go run main.go --atmos https://music.apple.com/us/album/album-name/id123456789
```

## 🎯 Ví dụ sử dụng

### Web Interface:
1. Mở http://localhost:8080
2. Dán URL Apple Music vào ô input
3. Chọn chất lượng âm thanh
4. Nhấn "Start Download"
5. Theo dõi tiến trình trong "Download Tasks"

### Command Line:
```bash
# Tải xuống album
go run main.go https://music.apple.com/us/album/1989-taylors-version/1713845538

# Tải xuống với Dolby Atmos
go run main.go --atmos https://music.apple.com/us/album/1989-taylors-version/1713845538

# Tải xuống AAC
go run main.go --aac https://music.apple.com/us/album/1989-taylors-version/1713845538

# Tìm kiếm
go run main.go --search album "taylor swift"
```

## 📁 Files được tải xuống

Files sẽ được lưu trong:
- **ALAC (Lossless)**: `AM-DL downloads/`
- **AAC**: `AM-DL-AAC downloads/`
- **Dolby Atmos**: `AM-DL-Atmos downloads/`

## 🔧 Troubleshooting nhanh

### Lỗi "Failed to get authorization token"
- Kiểm tra kết nối internet
- Đảm bảo `authorization-token` trong config.yaml đúng

### Lỗi "Media user token is required"
- Cần `media-user-token` để tải AAC-LC, lyrics và music videos
- Lấy token theo hướng dẫn ở trên

### Lỗi "Invalid Apple Music URL"
- Đảm bảo URL đúng định dạng Apple Music
- URL phải chứa: `/album/`, `/song/`, `/playlist/`, `/artist/`, `/music-video/`, `/station/`

## 📚 Tài liệu đầy đủ

- [README chính](README.md) - Hướng dẫn chi tiết
- [README Web Server](README-WEB.md) - Hướng dẫn Web Server
- [README CLI](README-CN.md) - Hướng dẫn Command Line

## ⚠️ Lưu ý quan trọng

- Chỉ sử dụng cho mục đích cá nhân
- Không chia sẻ tokens với người khác
- Tuân thủ điều khoản sử dụng của Apple Music
- Không phân phối lại nội dung đã tải xuống 