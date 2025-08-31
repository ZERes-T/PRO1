(function(){
  var NEW_D = 'mebelplace.com.kz';
  var OLD_D = 'zodak-test.ru';
  var host = location.hostname;

  // уважаем скрытие на 7 дней
  var hideUntil = +(localStorage.getItem('dn_hide_until')||0);
  if (Date.now() < hideUntil) return;

  var msg;
  if (host === OLD_D) {
    msg = '⚠️ Скоро переезжаем на ' + NEW_D + '.';
  } else if (host === NEW_D) {
    msg = '✅ Вы на новом домене ' + NEW_D + '.';
  } else {
    msg = 'ℹ️ Официальный домен: ' + NEW_D + '.';
  }

  function mount(){
    var header = document.querySelector('header');
    var bar = document.createElement('div');
    bar.className = 'ann';
    bar.innerHTML = '<div class="ann-i"><span>'+msg+'</span><span class="grow"></span>'
      + '<a class="btn-s" href="/domain.html">Подробнее</a>'
      + ' <button class="btn-s" id="annClose">✕</button></div>';
    if (header && header.parentNode) {
      header.parentNode.insertBefore(bar, header.nextSibling);
    } else {
      document.body.insertBefore(bar, document.body.firstChild);
    }
    var close = document.getElementById('annClose');
    if (close) close.onclick = function(){
      localStorage.setItem('dn_hide_until', Date.now() + 7*24*60*60*1000);
      bar.remove();
    };
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', mount);
  } else {
    mount();
  }
})();

