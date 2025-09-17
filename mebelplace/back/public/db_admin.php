<?php
// db_admin.php — временная страница управления БД

// ⚠️ ОБЯЗАТЕЛЬНО меняй доступ
$dbHost = "127.0.0.1";
$dbName = "zodak";
$dbUser = "zodak";
$dbPass = "Oljik_1872725";

try {
    $pdo = new PDO("pgsql:host=$dbHost;dbname=$dbName", $dbUser, $dbPass);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
} catch (PDOException $e) {
    die("Ошибка подключения: " . $e->getMessage());
}

// Получаем список таблиц
$tables = $pdo->query("SELECT tablename FROM pg_tables WHERE schemaname='public'")
              ->fetchAll(PDO::FETCH_COLUMN);

// Если выбрана таблица
$table = $_GET['table'] ?? null;
?>
<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8"/>
<title>DB Admin — MebelPlace</title>
<style>
 body{font-family:sans-serif;background:#0b0b0f;color:#e5e7eb;padding:20px}
 a{color:#f59e0b;text-decoration:none}
 table{border-collapse:collapse;margin-top:10px;width:100%;background:#11131a}
 th,td{border:1px solid #1f2430;padding:6px 10px}
 th{background:#1c1f2b}
 form{margin:10px 0}
 input,select,button{padding:6px 10px;border-radius:6px;border:1px solid #1f2430;background:#1c1f2b;color:#fff}
</style>
</head>
<body>
<h1>Управление БД</h1>

<h3>Таблицы:</h3>
<?php foreach($tables as $t): ?>
  <a href="?table=<?=htmlspecialchars($t)?>"><?=htmlspecialchars($t)?></a><br>
<?php endforeach; ?>

<?php if($table): ?>
  <h2>Таблица: <?=htmlspecialchars($table)?></h2>
  <?php
    $rows = $pdo->query("SELECT * FROM \"$table\" LIMIT 50")->fetchAll(PDO::FETCH_ASSOC);
    if($rows):
  ?>
    <table>
      <tr>
        <?php foreach(array_keys($rows[0]) as $col): ?>
          <th><?=htmlspecialchars($col)?></th>
        <?php endforeach; ?>
      </tr>
      <?php foreach($rows as $r): ?>
        <tr>
          <?php foreach($r as $val): ?>
            <td><?=htmlspecialchars($val)?></td>
          <?php endforeach; ?>
        </tr>
      <?php endforeach; ?>
    </table>
  <?php else: ?>
    <p>Нет записей</p>
  <?php endif; ?>
<?php endif; ?>
</body>
</html>
