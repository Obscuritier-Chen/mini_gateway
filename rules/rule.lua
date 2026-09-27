function check_request(req)
    local auth_token = req.headers["Authorization"] or ""
    local debug_mode = req.query["debug"] == "true"

    if debug_mode == true then
        print("[lua debug] Request ID: " ..(req.request_id or "" ).. "Method: ".. req.method)
    end

    if string.find(req.path, '^/admin') and auth_token ~= "Bearer secret123" then
        return 403, "Forbidden: Missing or invalid bearer token"
    end

    return 200, "OK"
end
