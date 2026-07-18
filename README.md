# Player Node

HTTP service หน้า embed player ของ [VdoHide](https://vdohide.xyz) — เสิร์ฟเฉพาะหน้า embed (content จริงทั้งหมด + player feeds อยู่ที่ content-node)

> แทนที่ `server-player` เดิม — **ไม่ใช่ worker** เป็น service ที่ต้องออนไลน์ตลอด (คนดูวิดีโอเปิด embed ผ่านตัวนี้)

## หน้าที่

1. **หน้า embed player**
   - `/embed/{fileSlug}` — หน้า player (HTML) ตรวจ custom domain / space / สถานะ processing
2. **Misc** — `/favicon.ico`, `/health`
   > ไม่ proxy content ใดๆ — stream / รูป / sprite / m3u8 / ซับ เป็นหน้าที่ของ **content-node**
   > `/playlist/{fileSlug}.json` และ `/advert/{adSlug}.json` **ย้ายไป content-node แล้ว** — static domain (`static.vdohide.org`) ต้อง route ไปที่ content-node
3. **Sync ทุก 1 นาที** — settings (`player_maintenance`, `advert_hobby`, `domain_*`), custom domains, spaces จาก MongoDB → เขียนลง `conf/*.json` ข้างๆ binary + cache ใน memory

## Requirements

- **MongoDB** (vdohide platform database)

---

## Installation (Linux Server)

```bash
curl -fsSL https://raw.githubusercontent.com/zergolf1994/player-node/main/install.sh | sudo -E bash -s -- \
    --database-url "mongodb+srv://user:pass@cluster.mongodb.net/platform"
```

| Option | Default | คำอธิบาย |
|---|---|---|
| `--database-url` | `""` | MongoDB connection string (`DATABASE_URL`) |
| `--port` | `8081` | HTTP port |
| `--domain-static` | `""` | static/content domain fallback (ปกติใช้ setting `domain_static` จาก DB) |
| `--uninstall` | — | ถอนการติดตั้ง |

```bash
journalctl -u player-node -f          # ดู logs
systemctl restart player-node         # restart
curl http://localhost:8081/health     # health check
```

## Configuration (.env)

```env
DATABASE_URL=mongodb+srv://user:pass@cluster.mongodb.net/platform
PORT=8081
DOMAIN_STATIC=
LOG_PATH=logs/player-node.log

# Optional — Redis lookup cache (ไม่ตั้ง = ไม่ใช้, Redis ล่ม = ยิง DB ตามปกติ)
# เก็บผล resolve ของหน้า embed: embed_resolve:{slug} (เฉพาะไฟล์ ready, TTL 300s)
REDIS_URL=redis://localhost:6379/0
```

## Development

```bash
go run ./cmd     # ต้องมี .env (DATABASE_URL)
build.bat        # Windows exe + copy .env → .build/
```

## Release

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions build + release อัตโนมัติ: `linux` (amd64), `linux-arm64`

---

## Collections Used

| Collection | การใช้งาน |
|---|---|
| `files` | lookup file ด้วย slug (embed) |
| `medias` | หา media (video/poster/thumbnail) ของ file + storageId |
| `storages` | หา publicUrl / host:port ของ storage สำหรับ proxy |
| `custom_domains` | sync ลง cache — ตรวจ domain / player config / adverts |
| `workspaces` | sync ลง cache — ตรวจ space status / plan |
| `settings` | sync `player_maintenance`, `advert_hobby`, `domain_*` |
| `video_process` | แสดงสถานะ processing/error บนหน้า embed ตอนวิดีโอยังไม่พร้อม |

> ⚠ **Index เป็นของฝั่ง vdohide-service (mongoose)** — repo นี้ไม่สร้าง index เอง
> ⚠ enum ใน `internal/core/enums/` ต้อง match กับ `vdohide-service/src/core/enums/`
