local function _normalize(value)
	local retval = ''
	if type(value) == 'function' then
		retval = '<' .. tostring(value) .. '>'
	elseif type(value) == 'table' then
		retval = '<' .. tostring(value) .. '>'
	elseif type(value) == 'string' then
		retval = string.format('%q',value)
	else
		retval = tostring(value)
	end
	return retval
end

local function _ToString(value, c)
	local rettbl = {}
	if type(value) == 'table' then
		c = (c or 0) + 1
		if c >= 100 then
			error("sys.dump too deep:")
			error(table.concat(rettbl))
		end

		table.insert(rettbl, '{')
		for k, v in pairs(value) do
			table.insert(rettbl, '[')
			table.insert(rettbl, _ToString(k,c))
			table.insert(rettbl, '] = ')
			table.insert(rettbl, _ToString(v,c))
			table.insert(rettbl, ', ')
		end
		table.insert(rettbl, '}')
	else
		table.insert(rettbl, _normalize(value))
	end
	return table.concat(rettbl)
end

function ToString(value)
	local ni, ret = pcall (_ToString, value)
	return ret
end

local REPR_INDENT = '  '
local function _ToStringIndent(value, forprint, deep)
	local linebr
	if forprint then
		linebr = '\n'
	else
		linebr = '#r'
	end
	local ret = ''
	if type(value) == 'table' then
		deep = deep or 0
		if deep >= 100 then error("too deep:"..ret) end
		local indent = string.rep(REPR_INDENT, deep) 
		ret = ret .. '{' .. linebr
		for k, v in pairs(value) do
			local krepr
			if type(k)=='string' and string.match(k, '[_%a][_%a%d]*')==k then
				krepr = k
			else
				krepr = '[' .. _ToStringIndent(k, forprint, deep + 1) .. ']'
			end
			ret = ret .. indent .. REPR_INDENT .. krepr .. ' = ' .. _ToStringIndent(v, forprint, deep + 1) .. ',' .. linebr
		end
		ret = ret .. indent .. '}'
		return ret 
	else
		return _normalize(value)
	end
end

function ToStringIndent(obj, forprint)
	local ni, ret = pcall(_ToStringIndent, obj, forprint)
	return ret
end

