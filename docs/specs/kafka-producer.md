# Спека: Kafka Producer

## Что делаем
Продюсер отправляет событие `OrderCreated` в топик `orders`.

## Входная структура
Order{ID, UserID, Amount, CreatedAt}

## Выходная структура
Kafka-сообщение: key = OrderID, value = JSON(OrderCreatedEvent)

## Бизнес-логика
- Валидация: Amount > 0, иначе не отправляем
- При ошибке отправки — retry 3 раза с backoff
- Логирование: structured log с trace_id

## Правила реализации
Следует `skills/kafka-adapter/SKILL.md` (сериализация, обработка ошибок, конфиг).