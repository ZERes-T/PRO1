<?php
header('Content-Type: application/json; charset=utf-8');

$dir  = realpath(__DIR__ . '/../media');
$base = '/media';

$items = [];
if ($dir && is_dir($dir)) {
  $dh = opendir($dir);
  while (($f = readdir($dh)) !== false) {
    if ($f[0] === '.') continue;
    $ext = strtolower(pathinfo($f, PATHINFO_EXTENSION));
    if (!in_array($ext, ['mp4','m3u8'])) continue;
    $name = pathinfo($f, PATHINFO_FILENAME);
    $items[] = [
      'id'     => crc32($f),
      'title'  => $name,
      'url'    => $base . '/' . $f,
      'poster' => $base . '/' . $name . '.jpg',
      'likes'  => 0, 'comments'=>0, 'shares'=>0
    ];
  }
  closedir($dh);
}
echo json_encode(['items'=>$items], JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES);
