<?php
// /api/requests.php
declare(strict_types=1);
header('Content-Type: application/json; charset=utf-8');

$root = rtrim($_SERVER['DOCUMENT_ROOT'] ?? dirname(__DIR__), '/');
$dbFile = $root . '/data/requests.sqlite';
@mkdir($root . '/data', 0775, true);
@mkdir($root . '/uploads', 0775, true);

try {
  $pdo = new PDO('sqlite:' . $dbFile, null, null, [
    PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
    PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
  ]);
  $pdo->exec("CREATE TABLE IF NOT EXISTS requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    height_cm REAL NOT NULL,
    width_cm  REAL NOT NULL,
    material  TEXT NOT NULL,
    description TEXT,
    photo_url TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  )");
} catch (Throwable $e) {
  http_response_code(500);
  echo json_encode(['ok'=>false,'error'=>'db']); exit;
}

if (($_SERVER['REQUEST_METHOD'] ?? 'GET') !== 'POST') {
  http_response_code(405); echo json_encode(['ok'=>false]); exit;
}

$h = (float)($_POST['height_cm'] ?? 0);
$w = (float)($_POST['width_cm'] ?? 0);
$m = trim((string)($_POST['material'] ?? 'др.'));
$d = trim((string)($_POST['description'] ?? ''));

if ($h<=0 || $w<=0) { http_response_code(400); echo json_encode(['ok'=>false,'error'=>'bad_dims']); exit; }

$photoUrl = null;
if (!empty($_FILES['photo']['tmp_name'])) {
  $ext = pathinfo($_FILES['photo']['name'], PATHINFO_EXTENSION);
  $name = 'req_' . time() . '_' . bin2hex(random_bytes(3)) . '.' . preg_replace('~[^a-z0-9]+~i','',$ext);
  $dst = $root . '/uploads/' . $name;
  if (move_uploaded_file($_FILES['photo']['tmp_name'], $dst)) {
    $photoUrl = '/uploads/' . $name;
  }
}

$st = $pdo->prepare("INSERT INTO requests (height_cm,width_cm,material,description,photo_url) VALUES (?,?,?,?,?)");
$st->execute([$h,$w,$m,$d,$photoUrl]);
$id = (int)$pdo->lastInsertId();

echo json_encode(['ok'=>true,'id'=>$id,'photo_url'=>$photoUrl], JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES);

