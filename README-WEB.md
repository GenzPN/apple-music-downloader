# Apple Music Downloader - Web Server

Ứng dụng web để tải xuống nhạc từ Apple Music với giao diện người dùng thân thiện.

## Tính năng

- 🌐 Giao diện web đẹp mắt và dễ sử dụng
- 📱 Responsive design, hoạt động tốt trên mobile
- 🎵 Hỗ trợ tải xuống nhiều loại nội dung:
  - Album
  - Bài hát đơn lẻ
  - Playlist
  - Nghệ sĩ (tất cả album)
  - Music Video
  - Radio Station
- 🎚️ Chọn chất lượng âm thanh:
  - Lossless (ALAC)
  - High-Quality (AAC)
  - Dolby Atmos
- 📊 Theo dõi tiến trình tải xuống real-time
- 📋 Quản lý nhiều task tải xuống cùng lúc
- 🔄 Tự động cập nhật trạng thái

## Cài đặt

### Yêu cầu hệ thống

- Go 1.19 hoặc cao hơn
- FFmpeg (cho animated artwork)
- MP4Box (từ GPAC)
- mp4decrypt (cho music videos)

### Cài đặt dependencies

```bash
# Cài đặt Go dependencies
go mod download

# Cài đặt FFmpeg (Ubuntu/Debian)
sudo apt update
sudo apt install ffmpeg

# Cài đặt GPAC (Ubuntu/Debian)
sudo apt install gpac

# Cài đặt mp4decrypt (từ Bento4)
# Tải từ: https://www.bento4.com/downloads/
```

### Cấu hình

1. Sao chép file `config.yaml` và chỉnh sửa:

```yaml
# Token cần thiết cho việc tải xuống AAC-LC, lyrics và music videos
media-user-token: "your-media-user-token"

# Token authorization (thường tự động lấy được)
authorization-token: "your-authorization-token"

# Ngôn ngữ
language: ""

# Storefront của tài khoản Apple Music (quan trọng!)
storefront: "us"  # Thay đổi theo quốc gia của bạn (us, jp, ca, vn, etc.)

# Các cài đặt khác...
```

### Lấy tokens

#### Media User Token
1. Mở Apple Music trên web browser
2. Mở Developer Tools (F12)
3. Vào tab Network
4. Tìm request đến `amp-api.music.apple.com`
5. Trong headers, tìm `media-user-token`

#### Authorization Token
Thường tự động lấy được, nhưng nếu cần:
1. Mở Apple Music trên web browser
2. Mở Developer Tools (F12)
3. Vào tab Network
4. Tìm request đến `amp-api.music.apple.com`
5. Trong headers, tìm `authorization`

## Sử dụng

### Chạy server

```bash
# Chạy với port mặc định (8080)
go run server_main.go server.go main.go

# Hoặc chạy với port tùy chỉnh
go run server_main.go server.go main.go -port 3000
```

### Truy cập web interface

Mở trình duyệt và truy cập: `http://localhost:8080`

### Sử dụng giao diện web

1. **Nhập URL Apple Music**: Dán URL từ Apple Music vào ô input
2. **Chọn chất lượng**: Chọn chất lượng âm thanh mong muốn
3. **Bắt đầu tải xuống**: Nhấn nút "Start Download"
4. **Theo dõi tiến trình**: Xem tiến trình trong phần "Download Tasks"

### Các loại URL được hỗ trợ

- **Album**: `https://music.apple.com/us/album/album-name/id123456789`
- **Song**: `https://music.apple.com/us/album/song-name/id123456789?i=987654321`
- **Playlist**: `https://music.apple.com/us/playlist/playlist-name/pl.123456789`
- **Artist**: `https://music.apple.com/us/artist/artist-name/id123456789`
- **Music Video**: `https://music.apple.com/us/music-video/video-name/id123456789`
- **Station**: `https://music.apple.com/us/station/station-name/ra.123456789`

## API Endpoints

### POST /api/download
Bắt đầu tải xuống

**Request Body:**
```json
{
  "url": "https://music.apple.com/us/album/...",
  "quality": "alac"
}
```

**Response:**
```json
{
  "task_id": "task_123456789",
  "status": "started"
}
```

### GET /api/status?task_id=task_id
Lấy trạng thái task

**Response:**
```json
{
  "id": "task_123456789",
  "url": "https://music.apple.com/...",
  "type": "album",
  "status": "processing",
  "progress": 50,
  "message": "Downloading album...",
  "created_at": "2024-01-01T12:00:00Z"
}
```

### GET /api/tasks
Lấy danh sách tất cả tasks

### GET /api/config
Lấy cấu hình hiện tại

## Cấu trúc thư mục tải xuống

Theo cấu hình trong `config.yaml`, files sẽ được tải xuống vào:

- **ALAC (Lossless)**: `AM-DL downloads/`
- **AAC**: `AM-DL-AAC downloads/`
- **Dolby Atmos**: `AM-DL-Atmos downloads/`

## Troubleshooting

### Lỗi "Failed to get authorization token"
- Kiểm tra kết nối internet
- Đảm bảo `authorization-token` trong config.yaml đúng

### Lỗi "Media user token is required"
- Cần `media-user-token` để tải AAC-LC, lyrics và music videos
- Lấy token theo hướng dẫn ở trên

### Lỗi "Invalid Apple Music URL"
- Đảm bảo URL đúng định dạng Apple Music
- URL phải chứa một trong các path: `/album/`, `/song/`, `/playlist/`, `/artist/`, `/music-video/`, `/station/`

### Lỗi "Failed to get lyrics"
- Kiểm tra `storefront` trong config.yaml có khớp với tài khoản Apple Music
- Đảm bảo `media-user-token` đúng

### Files không tải xuống
- Kiểm tra quyền ghi vào thư mục tải xuống
- Đảm bảo đủ dung lượng ổ cứng
- Kiểm tra logs trong terminal

## Tính năng nâng cao

### Tùy chỉnh format tên file
Chỉnh sửa trong `config.yaml`:

```yaml
# Format tên album
album-folder-format: "{ReleaseYear} - {ArtistName} - {AlbumName}"

# Format tên bài hát
song-file-format: "{SongNumer}. {SongName} [{Quality}]"

# Format tên nghệ sĩ
artist-folder-format: "{ArtistName}"
```

### Tải xuống lyrics
```yaml
embed-lrc: true          # Nhúng lyrics vào file
save-lrc-file: true      # Lưu lyrics thành file riêng
lrc-format: "lrc"        # Format: lrc hoặc ttml
```

### Tải xuống cover art
```yaml
embed-cover: true
cover-size: 5000x5000
cover-format: jpg        # jpg, png, hoặc original
```

## Bảo mật

⚠️ **Lưu ý quan trọng**: 
- Không chia sẻ tokens với người khác
- Chỉ sử dụng cho mục đích cá nhân
- Tuân thủ điều khoản sử dụng của Apple Music
- Không phân phối lại nội dung đã tải xuống

## Hỗ trợ

Nếu gặp vấn đề, hãy:
1. Kiểm tra logs trong terminal
2. Đảm bảo cấu hình đúng
3. Kiểm tra kết nối internet
4. Thử với URL khác để xác định vấn đề

## License

Dự án này chỉ dành cho mục đích giáo dục và sử dụng cá nhân. 