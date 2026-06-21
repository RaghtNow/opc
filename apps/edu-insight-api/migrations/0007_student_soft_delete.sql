UPDATE classroom_students
SET profile_status = 'archived',
    status = '已删除',
    parent_status = '已停用',
    selection_status = '已停用'
WHERE profile_status = 'deleted';
