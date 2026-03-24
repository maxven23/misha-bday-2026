# 🎂 birthday — CLI поздравление для Михаила Конькова

Интерактивное терминальное поздравление, написанное на Go с использованием [Charmbracelet](https://charm.sh/) (bubbletea, lipgloss, bubbles).

## Запуск

```bash
# Скачать из GitLab Releases и запустить
chmod +x birthday_linux_amd64
./birthday_linux_amd64

# Или собрать самому
go build -o birthday .
./birthday
```

**Управление:** `↑`/`↓` — прокрутка, `q` — выход.

## Сборка

```bash
go build .
```

## CI/CD

Пайплайн в `.gitlab-ci.yml`:
- `lint` — `go vet` + golangci-lint
- `build` — кросс-компиляция под linux/darwin/windows × amd64/arm64
- `release` — загрузка бинарей в Package Registry + создание Release (только по тегу)

Создать релиз:

```bash
git tag v1.0.0
git push origin v1.0.0
```
