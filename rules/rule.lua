function check_request(path, user_agent)

    if string.find(path, '^/admin') and user_agent ~= "Admin-Token" then
        return 403, "Forbidden: Invalid Admin Token"
    end

    return 200, "OK"
end