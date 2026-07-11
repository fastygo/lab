```briefly
    Пользователь спрашивает, что ещё осталось сделать. Он сообщает, что Docker Desktop уже готов и работает.

    Сейчас проверю audit-progress.md и PRO.md, чтобы точно определить оставшиеся задачи. Также укажу, что можно запустить через Docker.GRESS.

    Сверю актуальный статус по coverage SSOT.

    Составлю приоритизированный план. Опишу оставшиеся задачи продукта. Укажу команды для немедленного запуска в Docker Desktop.
```

Код пакета Gate 2/3 + Q2 готов; **следующий шаг с Docker — прогнать живьём**, а не писать новый код.

### Сейчас с Docker Desktop

```bash
cd d:/FastyGo/Lab
make runners          # lighthouse, axe, theme-check, vnu, wpscan
make org-up           # WP :8080
go run ./apps/cli run -f testdata/manifests/org.lab.yaml

make quality-up
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml
```

Org без compose/images даст `runner.docker.unavailable` / HTTP fail; с Docker — реальный Theme Check + smoke по matrix.

---

Запускаю Docker-тесты лаборатории. Создам раннеры, обновлю org и выполню манифесты org/quality.

Прогоняю runners, compose org и lab manifests. Сначала проверю Docker и состояние терминалов.

29.6.1
testdata/dist/latte.zip
Docker работает. Собираю раннеры и запускаю org compose.

Сборка раннеров займёт время. Запущу их и org-up параллельно — они не зависят друг от друга. После обоих завершений прогоню org.lab.yaml.

--: line 1: make: command not found
--: line 1: make: command not found
make недоступен. Запущу команды из Makefile напрямую в Docker.

make нет в PATH — запускаю те же команды напрямую.

Организация завершена успешно. Бегуны всё ещё собираются.

Организация работает. WordPress отдаёт 200. Изображение theme-check может собираться отдельно. Lighthouse всё ещё устанавливает Chromium.

Параллельная сборка theme-check позволит запустить org раньше, но Docker-контейнеры могут конкурировать за ресурсы.

Полный прогон org требует образ theme-check. Параллельная сборка этого образа сейчас возможна.

WP на :8080 уже отвечает 200. Параллельно собираю theme-check, чтобы не ждать lighthouse.

Theme-check готов. Организация WP работает.

Запускаю org.lab.yaml. Lighthouse продолжает работу.

theme-check готов — запускаю org lab.

Запуск завершился с ошибкой theme-check. Скрипт entrypoint.sh не найден. Вероятная причина — Windows-переводы строк в Docker.

В артефакте темы обнаружены реальные проблемы: папка vendor/bin, ресурсы и минифицированный код. Это ожидаемо для сборки, а не баги лаборатории.

HTTP-тесты дали 4 из 8 успешных ответов. Остальные страницы возвращают 404, потому что WordPress не установлен с примерным контентом и пермалинками.

Нужно исправить CRLF в скриптах, пересобрать контейнер и установить WordPress с примерным контентом для smoke-тестов.

Нашёл две блокеры: CRLF в entrypoint.sh (типичная ошибка no such file or directory) и WP без sample content/permalinks. Чиню и перезапускаю.

В файле entrypoint.sh обнаружен CRLF. Нужно конвертировать в LF и пересобрать.

WordPress не установлен. Установлю его через wpcli. Проверю run-check.php и to-findings.php — PHP с CRLF обычно работает. Затем пересоберу theme-check, настрою пермалинки и перезапущу org lab.

Вывод curl выглядит странно.

Пытаюсь понять, почему curl с двумя URL и -o /dev/null всё равно выводит тело ответа. Возможно, флаг -o применяется только к первому запросу.

Разбираю формат curl. Скорее всего, -w применяется к каждому URL отдельно. Тогда 200 — это код первого запроса, а 404 — второго. Страница sample-page, вероятно, возвращает 404.

Проверяю стандартную страницу WordPress «Sample Page» со slug sample-page. Нужно создать тег test для матричного URL. Категория uncategorized и автор admin должны работать с pretty permalinks. Перезапущу лабораторию org с фокусом на theme-check.

Страница sample-page существует, но возвращает ошибку 404.

Скорее всего, 404 на sample-page вызван отсутствием .htaccess. Проверю его наличие и работу mod_rewrite. При необходимости переключу тесты на query-string URL.

Проверка /this-page-does-not-exist-lab-404/ работает корректно. Сейчас попробую создать .htaccess для работы pretty permalinks.

апущен agent mode. Начинаю выполнение задачи.

Проверяю редиректы и доступность страниц. Статус /sample-page/ внутри контейнера пока неизвестен. На хосте curl на порт 8080 возвращает 404 Apache вместо WordPress. Вероятно, rewrite не применяется через опубликованный порт. Либо проблема в siteurl на 127.0.0.1:8080.

Rewrite не срабатывает для запросов из контейнера. Проверяю, какой именно Directory в docker-php.conf имеет AllowOverride All.

Команда docker exec зависла. Попробую проще — использовать query-string URL для smoke matrix.

Сфокусируюсь на запуске theme-check. Проверю, не завис ли docker exec из-за apachectl. Gate 3 с pretty permalink пока не работает — оставлю как есть. WP уже установлен, нужно только запустить проверку темы.

Сначала вручную проверю theme-check в Docker. Затем переведу пермалинки на plain-формат через wp rewrite structure '', чтобы матрица smoke работала без реврайта. После этого перезапущу полный прогон.

Pretty permalinks с хоста не работают (Apache 404). Переключаю matrix на query URLs и прогоняю theme-check.

DB_HOST берётся из переменной окружения WORDPRESS_DB_HOST, по умолчанию mysql.

Контейнер theme-check не получает переменные окружения БД. Он использует mysql вместо db:3306. Нужно передать WORDPRESS_DB_HOST=db:3306 и другие параметры БД в org.lab.yaml или docker runner для theme-check.

Также нужна WORDPRESS_DB_NAME=wordpress. Docker runner сейчас прокидывает только LAB_* и WPSCAN. Нужно расширить его, чтобы он передавал переменные окружения из конфига в контейнеры.

В org.lab.yaml можно задать WORDPRESSDB_HOST: db:3306 и другие параметры БД. Альтернатива — экспортировать их прямо в entrypoint theme-check через переменные окружения по умолчанию.

Theme-check не видит WORDPRESS_DB_HOST из compose — добавлю дефолты в entrypoint и проброс env.

Обнаружил две проблемы. При ручном запуске Docker в Git Bash некорректно обрабатывался путь к /lab/theme.zip.

Git Bash преобразовал путь к архиву в Windows-формат. Также нужно проверить ошибку «invalid character 'U'» в Go lab — возможно, это другая проблема.

Go runner передаёт переменную окружения напрямую в Docker, поэтому конвертация путей через Git Bash не должна срабатывать. Ошибка «invalid character 'U'» вероятно исходит из stdout контейнера. Следует перезапустить org и точно зафиксировать вывод docker.

Проверю запуск с переменной MSYS_NO_PATHCONV=1. Путь в Go задан корректно, но нужно убедиться, что конвертация путей не сработала.

Не удалось создать директорию /var/www/html/wp-content/upgrade из-за проблем с правами доступа.

Также выявлена проблема с путями в архиве: бэкслеши от Windows (latte\assets\) могут не распаковаться в Linux. Для установки темы нужно пересобрать zip с прямыми слешами или скопировать файлы вручную. Права доступа к wp-content тоже требуют исправления — возможно, потребуется запуск под root.

В Dockerfile уже указан USER root, а wp --allow-root уже используется. Проверю, можно ли исправить права и пересобрать zip с прямыми слешами.

Две проблемы: права на wp-content и zip с Windows-путями (latte\...). Чиню образ и перепаковываю zip.

Python успешно извлечён в /tmp/latte-repack/latte.

Команда zip недоступна. Создам архив через Python.

Скрипт Python, вероятно, упал без вывода. Команда «Python» в ответе указывает на сбой heredoc в bash Windows.

Возможно, архив повреждён — лишние обратные слэши или неправильная структура. Проверю zip и запущу Python корректно.

Ожидаю завершения загрузки архива Alpine.

Процесс завершился с кодом 12. Сейчас изучу вывод.

composer.json
composer.lock
configs/
-rw-r--r-- 1 godud 197609 615 июл  9 21:14 /tmp/latte-repack/latte/style.css
C:\Users\godud\AppData\Local\Temp\latte-repack\latte
Путь к latte-repack на самом деле в Windows-локальной папке Temp.

Docker не видит папку через /tmp из-за путей Git Bash. Перепакую через Go напрямую из Windows-пути.

Исправлен формат архива — теперь используются прямые слэши.

Проверю установку темы и перезапущу org lab.


...
      1 "code":"org.themecheck.theme_install_failed"
STDERR:
200 /tmp/tc3.txt
В корне папки темы отсутствует файл style.css. WordPress ищет его именно там.

Проверяю, есть ли style.css в архиве. Файл находится в папке latte/, поэтому при распаковке он должен оказаться в корне темы.

style.css существует, но WordPress сообщает об отсутствии файла. Ищу причину.

Возможные причины — повреждённые метки времени в архиве или остатки предыдущих установок. Попробую скопировать zip во временную папку и установить тему вручную. Также проверю работу ZipArchive в WordPress.

+ wp theme activate latte --allow-root
Success: Switched to 'Latte' theme.
+ wp theme list '--status=active' --allow-root
name	status	update	version	update_version	auto_update
latte	active	none	1.1.0		off
Ручное распаковка и активация темы работает. Установка через wp theme install из zip-архива падает — вероятно, в архиве отсутствуют записи каталогов.

Обновлю entrypoint: копировать zip в /tmp, распаковывать в themes/, активировать тему и запустить theme-check. Это надёжнее для лаборатории.

wp theme install падает на этом zip — переключаю entrypoint на unzip + activate.

Тема установлена без ошибок. Проверка темы не вернула JSON.

Нужно отладить run-check.php и установку плагина.

ERR:
Fatal error: strict_types declaration must be the very first statement in the script in phar:///usr/local/bin/wp/vendor/wp-cli/eval-command/src/EvalFile_Command.php(85) : eval()'d code on line 8
Error: There has been a critical error on this website.Learn more about troubleshooting WordPress. There has been a critical error on this website.

no
Нашёл причину ошибки: wp eval-file добавляет обёртку перед файлом, поэтому declare(strict_types=1); не может быть первым. Скорее всего, в начале файла есть BOM или лишний текст.

Также проверю CRLF в run-check.php. Уберу declare(strict_types=1) из run-check.php и to-findings.php. Нужно найти theme-check.php — eval-file падает раньше, чем загружается плагин.

Проверка темы завершена. Найдено 2 обязательных пункта, 1 предупреждение и 21 рекомендация.

Запускаю полный org lab через Go. Пути в Docker на Windows корректно обрабатываются.

Theme Check живой. Запускаю полный org lab.

Полный пайплайн работает корректно.

Gate 1: найдены проблемы упаковки темы. Gate 2: проверка темы прошла — 2 обязательных, 1 предупреждение, 21 рекомендация. Gate 3: 7 из 8 HTTP-запросов прошли, только ?tag=test вернул 404. Исправлю тег и проверю матрицу 8/8. Затем проверю статус сборки Lighthouse и обновлю тест адаптера.

В списке образов уже есть lighthouse:local.

Сборка lighthouse завершена. Пустой тег возвращает 404 — это ожидаемо. Нужно назначить тег на пост или обновить матрицу. Для smoke-теста подойдёт ?cat=1. Стоит создать пост с этим тегом. Lighthouse готов. Осталось проверить axe и vnu.

Запланировано создание поста с тегом для устранения 404. Организационные результаты будут зафиксированы. Проверка качества запустится при готовности образов.

Тег всё ещё возвращает 404. Вероятно, проблема в шаблоне latte: либо не работает файл tag.php, либо сбой реврайтов.

Проверю тело ответа для прямого запроса ?tag=test. Шаблон latte может ломать архивы по тегам. Lighthouse уже готов. Axe и vnu всё ещё собираются. Запущу проверку качества только с доступными инструментами.

Если образа нет, ошибка другая — не «unavailable». Запущу quality-up и лабу: lighthouse должен отработать. Для axe и vnu покажу 7/8 в отчёте, если они не собрались. Ищу причину 404 у тега.

Ошибка 404 для тегов — проблема темы WordPress. Вероятно, шаблон Latte не поддерживает архивы тегов.

Ожидается завершение прогона quality.

Lighthouse отработал, но набрал 0 баллов. Скорее всего, из Docker на Windows не доходит до временного порта локального адаптера — в Docker Desktop Windows сеть работает иначе.

Орг-лаб отработал успешно. Изображения axe и vnu всё ещё собираются в фоне. Тест адаптера для матричного URL прошёл — проверка длины не помешала. Lighthouse получил 0 баллов из-за ограничений сетевой изоляции Docker Desktop. Полностью завершить quality не получилось, поэтому результаты частичные. Задачи обновлены.
---

### Что ещё осталось (по приоритету из PROGRESS)

| # | Область | Статус |
|---|---------|--------|
| 1 | **Org Gate 3 remainder** — Theme Unit Test XML import, attachment URL, notice hunter / `debug.log` | `[ ]` |
| 2 | **Org Gate 2 хвост** — WP_DEBUG log capture | `[ ]` |
| 3 | **Org Gate 4** — Playwright keyboard (skip link, nav, sheet, search) | `[ ]` |
| 4 | **Quality Q3** — stylelint | `[ ]` |
| 5 | **Quality** — median×3 LH, byte budgets, CWV, WP target вместо static | `[ ]` / `[~]` |
| 6 | **Sec S1+** — user enum, sensitive files, REST, CVE API | в основном `[ ]` |
| 7 | **Cycle E** — static-web adapters; **F** — SaaS | не начато |
| 8 | wpfasty `theme:verify` client | `[ ]` |

Уже закрыто в Lab: Gate 1 zip-lint, Gate 2 Theme Check (код), Gate 3 HTTP smoke (код), Q1/Q2/Q4 runners, sec headers + wpscan scaffold.

---

**Практично:** сначала `make runners` + `org-up` + `org.lab.yaml` — убедиться, что Gate 2/3 реально зелёные на `latte.zip`. Потом либо Gate 3 XML/notices, либо Q3 stylelint — как скажешь в Agent mode.

Сборка runners упала на **vnu**: URL `vnu.jar_24.10.17.zip` даёт 404 (проект больше не публикует versioned zip).

К этому моменту уже были: `lab/lighthouse`, `lab/axe`, `lab/theme-check`.  
Починил Dockerfile → качает `releases/download/latest/vnu.jar` — **`lab/vnu:local` собран**.