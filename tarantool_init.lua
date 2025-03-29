-- Инициализация Tarantool

-- Настройка пользователя
box.cfg {
    listen = 3301,
    memtx_memory = 128 * 1024 * 1024, -- 128MB
}

-- Создание пользователя admin с паролем secret и предоставление прав
box.schema.user.create('admin', {password = 'secret', if_not_exists = true})
box.schema.user.grant('admin', 'read,write,execute', 'universe', nil, {if_not_exists = true})
box.schema.user.passwd('admin', 'secret')

-- Создание пространства polls
local polls = box.schema.space.create('polls', {if_not_exists = true})

-- Создание первичного индекса (по ID)
polls:create_index('primary', {parts = {{1, 'string'}}, if_not_exists = true})

-- Функция для голосования
function vote_func(poll_id, option)
    local poll = box.space.polls:get(poll_id)
    if poll == nil then
        return {success = false, error = "poll not found"}
    end
    
    if poll[6] == true then -- проверка на завершенность голосования
        return {success = false, error = "poll is finished"}
    end
    
    local votes = poll[5]
    if votes[option] ~= nil then
        votes[option] = votes[option] + 1
        box.space.polls:update(poll_id, {{'=', 5, votes}})
        return {success = true}
    else
        return {success = false, error = "option not found"}
    end
end

-- Функция для завершения голосования
function finish_poll(poll_id)
    local poll = box.space.polls:get(poll_id)
    if poll == nil then
        return {success = false, error = "poll not found"}
    end
    
    box.space.polls:update(poll_id, {{'=', 6, true}})
    return {success = true}
end

-- Функция для удаления голосования
function delete_poll(poll_id)
    local poll = box.space.polls:get(poll_id)
    if poll == nil then
        return {success = false, error = "poll not found"}
    end
    
    box.space.polls:delete(poll_id)
    return {success = true}
end

print("Tarantool initialized successfully")
