# ベースイメージを指定
# ベースイメージを指定
FROM public.ecr.aws/docker/library/golang:1.18 as builder

# ワーキングディレクトリを設定
WORKDIR /app

# Goの依存関係を管理するためのファイルをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# 実行可能ファイルをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# 実行ステージ
FROM public.ecr.aws/docker/library/alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# ビルドした実行ファイルをコピー
COPY --from=builder /app/server .

# ポート80を開放
EXPOSE 80

# 実行可能ファイルを実行
CMD ["./server"]
