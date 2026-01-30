<?php
session_start();

// Initialize tasks array if it doesn't exist
if (!isset($_SESSION['tasks'])) {
    $_SESSION['tasks'] = [];
}

// Handle form submissions
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (isset($_POST['action'])) {
        switch ($_POST['action']) {
            case 'add':
                if (!empty($_POST['task'])) {
                    $_SESSION['tasks'][] = [
                        'id' => uniqid(),
                        'text' => htmlspecialchars($_POST['task']),
                        'completed' => false,
                        'created_at' => date('Y-m-d H:i:s')
                    ];
                }
                break;
            case 'delete':
                if (isset($_POST['task_id'])) {
                    $_SESSION['tasks'] = array_filter($_SESSION['tasks'], function($task) {
                        return $task['id'] !== $_POST['task_id'];
                    });
                }
                break;
            case 'toggle':
                if (isset($_POST['task_id'])) {
                    foreach ($_SESSION['tasks'] as &$task) {
                        if ($task['id'] === $_POST['task_id']) {
                            $task['completed'] = !$task['completed'];
                            break;
                        }
                    }
                }
                break;
            case 'clear':
                $_SESSION['tasks'] = [];
                break;
        }
    }
    // Redirect to prevent form resubmission
    header('Location: ' . $_SERVER['PHP_SELF']);
    exit;
}

$taskCount = count($_SESSION['tasks']);
$completedCount = count(array_filter($_SESSION['tasks'], function($task) {
    return $task['completed'];
}));
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PHP Task Manager</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 15px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            width: 100%;
            max-width: 600px;
            padding: 30px;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 2em;
        }
        .stats {
            color: #666;
            margin-bottom: 20px;
            font-size: 0.9em;
        }
        .stats span {
            display: inline-block;
            margin-right: 15px;
            padding: 5px 10px;
            background: #f0f0f0;
            border-radius: 5px;
        }
        .add-task-form {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        input[type="text"] {
            flex: 1;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 1em;
            transition: border-color 0.3s;
        }
        input[type="text"]:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            padding: 12px 20px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 1em;
            transition: background 0.3s;
        }
        button:hover {
            background: #5568d3;
        }
        .clear-btn {
            background: #e74c3c;
            width: 100%;
            margin-top: 10px;
        }
        .clear-btn:hover {
            background: #c0392b;
        }
        .tasks-list {
            list-style: none;
        }
        .task-item {
            background: #f8f9fa;
            padding: 15px;
            margin-bottom: 10px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            gap: 10px;
            transition: all 0.3s;
        }
        .task-item:hover {
            background: #e9ecef;
        }
        .task-item.completed {
            opacity: 0.6;
        }
        .task-item.completed .task-text {
            text-decoration: line-through;
            color: #999;
        }
        .task-text {
            flex: 1;
            word-break: break-word;
        }
        .task-meta {
            font-size: 0.8em;
            color: #999;
            display: block;
            margin-top: 5px;
        }
        .task-btn {
            padding: 8px 12px;
            font-size: 0.9em;
        }
        .toggle-btn {
            background: #27ae60;
        }
        .toggle-btn:hover {
            background: #229954;
        }
        .delete-btn {
            background: #e74c3c;
        }
        .delete-btn:hover {
            background: #c0392b;
        }
        .empty-state {
            text-align: center;
            padding: 40px;
            color: #999;
        }
        .server-info {
            margin-top: 20px;
            padding: 15px;
            background: #f0f0f0;
            border-radius: 8px;
            font-size: 0.85em;
            color: #666;
        }
        .server-info strong {
            color: #333;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>PHP Task Manager</h1>
        <div class="stats">
            <span>Total: <?php echo $taskCount; ?></span>
            <span>Completed: <?php echo $completedCount; ?></span>
            <span>Pending: <?php echo $taskCount - $completedCount; ?></span>
        </div>

        <form method="POST" class="add-task-form">
            <input type="hidden" name="action" value="add">
            <input type="text" name="task" placeholder="Enter a new task..." required>
            <button type="submit">Add Task</button>
        </form>

        <?php if (empty($_SESSION['tasks'])): ?>
            <div class="empty-state">
                <p>No tasks yet. Add one to get started!</p>
            </div>
        <?php else: ?>
            <ul class="tasks-list">
                <?php foreach ($_SESSION['tasks'] as $task): ?>
                    <li class="task-item <?php echo $task['completed'] ? 'completed' : ''; ?>">
                        <div class="task-text">
                            <?php echo $task['text']; ?>
                            <span class="task-meta">Created: <?php echo $task['created_at']; ?></span>
                        </div>
                        <form method="POST" style="display: inline;">
                            <input type="hidden" name="action" value="toggle">
                            <input type="hidden" name="task_id" value="<?php echo $task['id']; ?>">
                            <button type="submit" class="task-btn toggle-btn">
                                <?php echo $task['completed'] ? 'Undo' : 'Done'; ?>
                            </button>
                        </form>
                        <form method="POST" style="display: inline;">
                            <input type="hidden" name="action" value="delete">
                            <input type="hidden" name="task_id" value="<?php echo $task['id']; ?>">
                            <button type="submit" class="task-btn delete-btn">Delete</button>
                        </form>
                    </li>
                <?php endforeach; ?>
            </ul>

            <form method="POST">
                <input type="hidden" name="action" value="clear">
                <button type="submit" class="clear-btn" onclick="return confirm('Are you sure you want to clear all tasks?')">
                    Clear All Tasks
                </button>
            </form>
        <?php endif; ?>

        <div class="server-info">
            <strong>Server Information:</strong><br>
            PHP Version: <?php echo phpversion(); ?><br>
            Server: <?php echo $_SERVER['SERVER_SOFTWARE'] ?? 'Unknown'; ?><br>
            Session ID: <?php echo session_id(); ?><br>
            Request Time: <?php echo date('Y-m-d H:i:s'); ?>
        </div>
    </div>
</body>
</html>
