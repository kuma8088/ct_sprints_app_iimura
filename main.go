package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func init() {
	// .envファイルから環境変数を読み込む
	godotenv.Load()
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーを設定
	w.Header().Set("Access-Control-Allow-Origin", "*")             // すべてのオリジンからのアクセスを許可
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS") // GETとOPTIONSメソッドを許可
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // 特定のヘッダーの使用を許可

	// リクエストメソッドがOPTIONSの場合は、プリフライトリクエストとして扱う
	if r.Method == "OPTIONS" {
		return // プリフライトリクエストにはステータス200で応答して、処理を終了する
	}

	// hello worldという文字列をレスポンスとして返す
	fmt.Fprintf(w, "API接続テストが成功しました")
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーを設定
	w.Header().Set("Access-Control-Allow-Origin", "*")             // すべてのオリジンからのアクセスを許可
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS") // GETとOPTIONSメソッドを許可
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // 特定のヘッダーの使用を許可

	// リクエストメソッドがOPTIONSの場合は、プリフライトリクエストとして扱う
	if r.Method == "OPTIONS" {
		return // プリフライトリクエストにはステータス200で応答して、処理を終了する
	}

	reservation_count, err := database_test()

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 正しいフォーマットでレスポンスを返す
	fmt.Fprintf(w, "データベース接続テストが成功しました（Reservationsの件数：%d）", reservation_count)
}

func database_test() (int, error) {
	// 環境変数からデータベース接続の各要素を取得
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	servername := os.Getenv("DB_SERVERNAME")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	// 接続文字列を組み立て
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, servername, port, dbname)
	if connectionString == "" {
		log.Fatal("DB connection string is not set")
	}

	// データベースに接続
	connection, err := sql.Open(
		"mysql",
		connectionString)
	if err != nil {
		return 0, err
	}
	defer connection.Close()

	// SQLの実行
	rows, err := connection.Query("SELECT COUNT(*) AS reservation_count FROM Reservations")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// 結果の読み取り
	var reservation_count int
	for rows.Next() {
		err := rows.Scan(&reservation_count)
		if err != nil {
			return 0, err
		}
	}

	return reservation_count, nil
}

func initDB() error {
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	servername := os.Getenv("DB_SERVERNAME")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	// 1. DB作成のためにDB名なしで接続
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, servername, port)
	db, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to mysql root: %w", err)
	}
	defer db.Close()

	// 2. データベース作成
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// 3. テーブル作成のためにDB名ありで接続
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, servername, port, dbname)
	db2, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db2.Close()

	// テーブル作成
	_, err = db2.Exec(`CREATE TABLE IF NOT EXISTS Reservations (
		ID INT AUTO_INCREMENT PRIMARY KEY,
		company_name VARCHAR(255) NOT NULL,
		reservation_date DATE NOT NULL,
		number_of_people INT NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// サンプルデータ投入
	_, err = db2.Exec(`INSERT INTO Reservations (company_name, reservation_date, number_of_people)
		SELECT '株式会社テスト', '2024-04-21', 5
		WHERE NOT EXISTS (SELECT 1 FROM Reservations WHERE company_name = '株式会社テスト' AND reservation_date = '2024-04-21' AND number_of_people = 5)`)
	if err != nil {
		return fmt.Errorf("failed to insert sample data: %w", err)
	}

	return nil
}

func main() {
	apiport := os.Getenv("API_PORT")
	if apiport == "" {
		apiport = "8080"
	}

	// DB初期化
	if err := initDB(); err != nil {
		log.Printf("DB Initialization failed: %v", err)
		// 初期化失敗しても起動はする（ログで気付けるように）
	} else {
		log.Println("DB Initialization successful")
	}

	// /パスにアクセスがあった場合に、helloHandler関数を実行するように設定
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/test", TestHandler)

	// 8080ポートでサーバーを起動
	fmt.Println("HTTPサーバを起動しました。ポート: " + apiport)
	err := http.ListenAndServe(":"+apiport, nil)
	if err != nil {
		fmt.Println("HTTPサーバの起動に失敗しました: ", err)
	}

}
