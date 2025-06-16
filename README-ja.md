<div align="center">
  <a name="readme-top"></a>
  <h1><a href="https://www.billionmail.com/" target="_blank">BillionMail 📧</a></h1>


## スマートなキャンペーンのためのオープンソースメールサーバー／メールマガジン／Eメールマーケティングソリューション

[![][license-shield]][license-link] [![][docs-shield]][docs-link] [![][github-release-shield]][github-release-link] [![][github-stars-shield]][github-stars-link]

[English](README.md) | [简体中文](README-zh_CN.md) | 日本語
</div>
<br/>

<div align="center">
<a href="https://trendshift.io/repositories/13842" target="_blank"><img src="https://trendshift.io/api/badge/repositories/13842" alt="aaPanel%2FBillionMail | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>
</div>

## BillionMailとは？

BillionMailは、ビジネスや個人がメールキャンペーンを簡単に管理できるよう設計された**オープンソースのメールサーバー兼Eメールマーケティングプラットフォーム**です。ニュースレター、プロモーションメール、取引通知などを送信する際に、メールマーケティングのすべてを**完全にコントロール**できます。**高度な分析機能**や**顧客管理機能**を活用し、プロフェッショナルのようにメールを作成、送信、トラッキングできます。

![BillionMailバナー](https://www.billionmail.com/home.png?v1)

# たった3ステップで10億通のメールを送信！
**10億通のメール。あらゆるビジネスに。保証付き。**

### Step 1️⃣ BillionMailのインストール 
✅ インストールから**8分**で**✅ メール送信成功**まで完了します
```shell
cd /opt && git clone https://github.com/aaPanel/BillionMail && cd BillionMail && bash install.sh
````

### Step 2️⃣ ドメインを接続する

* 送信ドメインを追加
* DNSレコードを検証
* 無料SSLを自動有効化

### Step 3️⃣ キャンペーンを構築する

* メールを作成または貼り付け
* リストとタグを選択
* 送信日時を設定または今すぐ送信

<div align="center">
  <a href="https://www.youtube.com/embed/UHgxZa_9jGs?si=0-f1B5hDtcWImvQv" target="_blank">
    <img src="https://img.youtube.com/vi/UHgxZa_9jGs/maxresdefault.jpg" alt="" width="80%">
    <br />
    <img src="https://www.iconfinder.com/icons/317714/download/png/16" alt="YouTube" width="16"/>
    <b>YouTubeで視聴</b>
  </a>
</div>

## その他のインストール方法

👉 [https://www.aapanel.com/new/download.html](https://www.aapanel.com/new/download.html)

### aaPanelでワンクリックインストール

**Docker**

```shell
cd /opt && git clone https://github.com/aaPanel/BillionMail && cd BillionMail && cp env_init .env && docker compose up -d || docker-compose up -d
```

## 管理スクリプト

* 管理ヘルプ

  `bm help`

* デフォルトログイン情報を表示

  `bm default`

* ドメインのDNSレコードを表示

  `bm show-record`

* BillionMailを更新

  `bm update`

## ライブデモ

BillionMailデモ: [https://demo.billionmail.com/billionmail](https://demo.billionmail.com/billionmail)

ユーザー名: `billionmail`

パスワード: `billionmail`

## Webメール

BillionMailには**RoundCube**が統合されており、`/roundcube/`からWebメールにアクセスできます。

## なぜBillionMailを選ぶのか？

ほとんどのEメールマーケティングプラットフォームは**高価**、**クローズドソース**、または**基本機能が不足**しています。BillionMailはこれらと異なります：

✅ **完全オープンソース** – 隠れたコストなし、ベンダーロックインなし。
📊 **高度な分析機能** – メール配信、開封率、クリック率などを追跡。
📧 **送信数無制限** – 送信メール数に制限なし。
🎨 **カスタマイズ可能なテンプレート** – プロフェッショナルなマーケティングテンプレートを再利用可能。
🔒 **プライバシーファースト** – データは自分のサーバーにあり、サードパーティによる追跡なし。
🚀 **セルフホスト** – 自分のサーバーで実行し、完全にコントロール可能。

## どうすれば貢献できるか 🌟

BillionMailは**コミュニティ主導のプロジェクト**であり、立ち上げには皆さんのサポートが必要です！以下の方法でご参加ください：

1. **このリポジトリにスターを付ける**：スターを付けて関心を示しましょう。
2. **情報を拡散する**：開発者、マーケター、オープンソース愛好家にBillionMailを紹介しましょう。
3. **フィードバックを共有する**：Issueを立てるかディスカッションに参加して、どんな機能がほしいか教えてください。
4. **コントリビュートする**：開発が始まったら、コミュニティからの貢献を歓迎します。今後のアップデートをお待ちください！

---

📧 **BillionMail – オープンソースEメールマーケティングの未来。**

## Issues

問題が発生したり機能リクエストがある場合は、[Issueを作成](https://github.com/aaPanel/BillionMail/issues)してください。以下を含めると助かります：

* 問題またはリクエストの明確な説明
* 再現手順（該当する場合）
* スクリーンショットやエラーログ（該当する場合）

## 今すぐインストール

✅ インストールから**8分**で**メール送信成功**まで完了します

```shell
cd /opt && git clone https://github.com/aaPanel/BillionMail && cd BillionMail && bash install.sh
```

**Dockerでインストール:**（Dockerとdocker-compose-pluginを手動でインストールし、.envファイルを編集してください）

```shell
cd /opt && git clone https://github.com/aaPanel/BillionMail && cd BillionMail && cp env_init .env && docker compose up -d || docker-compose up -d
```

## ライセンス

BillionMailは**AGPLv3ライセンス**のもとで公開されています。これにより以下が可能です：

✅ ソフトウェアを無料で使用する
✅ コードを改変・再配布する
✅ プライベート利用に制限なし

詳細は[LICENSE](LICENSE)ファイルをご覧ください。

---

<!-- BillionMail公式リンク -->

[docs-link]: https://www.billionmail.com/

<!-- その他のリンク -->
[license-link]: https://www.gnu.org/licenses/gpl-3.0.html
[github-release-link]: https://github.com/aaPanel/BillionMail/releases/latest
[github-stars-link]: https://github.com/aaPanel/BillionMail
[github-issues-link]: https://github.com/aaPanel/BillionMail/issues

<!-- シールドリンク -->
[docs-shield]: https://img.shields.io/badge/documentation-148F76
[github-release-shield]: https://img.shields.io/github/v/release/aaPanel/BillionMail
[github-stars-shield]: https://img.shields.io/github/stars/aaPanel/BillionMail?color=%231890FF&style=flat-square
[license-shield]: https://img.shields.io/github/license/aaPanel/BillionMail
