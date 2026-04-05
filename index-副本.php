<?php
/**
 * 夸克网盘转存独立API接口（支持递归广告过滤）
 * 
 * 使用方法：
 * 1. 配置下方的COOKIE和资源目录ID
 * 2. 访问: kuake.php?url=https://pan.quark.cn/s/xxxxx&code=提取码(可选)&expired_type=1
 * 
 * 参数说明：
 * - url: 夸克分享链接（必填）
 * - code: 提取码（可选，如果分享链接需要提取码）
 * - expired_type: 1=永久资源 2=临时资源（默认1）
 * - ad_fid: 分享时额外带上的广告文件ID（可选）
 * 
 * 返回JSON格式：
 * {
 *   "code": 200,
 *   "message": "转存成功",
 *   "data": {
 *     "share_url": "分享链接",
 *     "code": "提取码",
 *     "title": "资源标题",
 *     "fid": "文件ID",
 *     "execution_time": "执行耗时(秒)",
 *     "filter_log": {
 *       "scan_start_time": "扫描开始时间",
 *       "scan_end_time": "扫描结束时间",
 *       "structure": ["文件结构树形显示"],
 *       "total_files": "总文件数",
 *       "total_folders": "总文件夹数",
 *       "ad_files": ["广告文件列表"],
 *       "deleted_fids": ["已删除的文件ID"],
 *       "result": "处理结果"
 *     }
 *   }
 * }
 * 
 * 广告过滤功能说明：
 * - 递归扫描所有层级的文件和文件夹
 * - 删除文件名或文件夹名包含广告关键词的项目
 * - 如果所有文件都是广告，返回"资源内容为空"错误
 * - 如果部分文件是广告，只删除广告文件，保留正常文件
 * - 输出详细的文件结构树和删除日志
 */

// ==================== 配置区域 ====================
// 夸克网盘Cookie（必填）- 登录夸克网盘后从浏览器复制Cookie
define('QUARK_COOKIE', ' ');

// 永久资源存储目录ID（必填）- 在夸克网盘创建文件夹后，获取其fid
define('QUARK_FILE_DIR', '0'); // 0表示根目录

// 临时资源存储目录ID（可选）- 用于临时资源
define('QUARK_FILE_TIME_DIR', '0');

// 广告关键词黑名单（可选）- 逗号分隔，转存后递归扫描所有子文件夹并删除包含这些关键词的文件/文件夹
define('QUARK_BANNED_KEYWORDS', '广告,上,福利');

// 是否启用广告过滤（true/false）- 启用后会递归扫描所有层级并输出详细日志
define('ENABLE_AD_FILTER', true);

// API请求超时时间（秒）
define('API_TIMEOUT', 60);

// 任务轮询最大次数
define('MAX_RETRY_COUNT', 50);
// ==================== 配置区域结束 ====================

// 记录开始时间
$start_time = microtime(true);

// 设置响应头
header('Content-Type: application/json; charset=utf-8');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST');

// 检查配置
if (QUARK_COOKIE === '你的夸克Cookie' || empty(QUARK_COOKIE)) {
    outputJson(500, '请先配置QUARK_COOKIE');
}

// 获取URL参数
$share_url = $_REQUEST['url'] ?? '';
$passcode = $_REQUEST['code'] ?? '';
$expired_type = intval($_REQUEST['expired_type'] ?? 1);
$ad_fid = $_REQUEST['ad_fid'] ?? '';

// 验证参数
if (empty($share_url)) {
    outputJson(400, '参数url不能为空');
}

// 解析分享链接
$pwd_id = parseShareUrl($share_url);
if (!$pwd_id) {
    outputJson(400, '资源地址格式有误，请提供正确的夸克分享链接');
}

// 执行转存
try {
    $result = transferQuarkResource($pwd_id, $passcode, $expired_type, $ad_fid);
    
    // 计算执行时间
    $execution_time = round(microtime(true) - $start_time, 2);
    $result['execution_time'] = $execution_time . '秒';
    
    // 添加广告过滤日志
    global $filterLog;
    if (!empty($filterLog)) {
        $result['filter_log'] = $filterLog;
    }
    
    outputJson(200, '转存成功', $result);
} catch (Exception $e) {
    // 即使出错，也返回过滤日志（如果有的话）
    global $filterLog;
    $errorData = [];
    if (!empty($filterLog)) {
        $errorData['filter_log'] = $filterLog;
    }
    outputJson(500, $e->getMessage(), !empty($errorData) ? $errorData : null);
}

/**
 * 解析夸克分享链接
 */
function parseShareUrl($url) {
    $substring = strstr($url, 's/');
    if ($substring === false) {
        return false;
    }
    $pwd_id = substr($substring, 2);
    // 去除可能的锚点
    $pwd_id = strtok($pwd_id, '#');
    return $pwd_id;
}

/**
 * 夸克资源转存主函数
 */
function transferQuarkResource($pwd_id, $passcode, $expired_type, $ad_fid) {
    // 步骤1: 获取stoken
    $tokenRes = getStoken($pwd_id, $passcode);
    if ($tokenRes['status'] !== 200) {
        throw new Exception($tokenRes['message'] ?? '获取访问令牌失败');
    }
    $stoken = str_replace(' ', '+', $tokenRes['data']['stoken']);
    
    // 步骤2: 获取分享资源详情
    $detailRes = getShareDetail($pwd_id, $stoken);
    if ($detailRes['status'] !== 200) {
        throw new Exception($detailRes['message'] ?? '获取资源详情失败');
    }
    
    $detail = $detailRes['data'];
    $title = $detail['share']['title'] ?? '未知资源';
    
    $fid_list = [];
    $fid_token_list = [];
    foreach ($detail['list'] as $item) {
        $fid_list[] = $item['fid'];
        $fid_token_list[] = $item['share_fid_token'];
    }
    
    if (empty($fid_list)) {
        throw new Exception('资源列表为空');
    }
    
    // 步骤3: 转存资源到自己网盘
    $to_pdir_fid = ($expired_type == 2) ? QUARK_FILE_TIME_DIR : QUARK_FILE_DIR;
    $saveRes = saveShareToMyDrive($pwd_id, $stoken, $fid_list, $fid_token_list, $to_pdir_fid);
    if ($saveRes['status'] !== 200) {
        throw new Exception($saveRes['message'] ?? '转存失败');
    }
    $task_id = $saveRes['data']['task_id'];
    
    // 步骤4: 轮询转存任务状态
    $saveData = pollTask($task_id);
    if (!$saveData || $saveData['status'] != 2) {
        throw new Exception('转存任务超时或失败');
    }
    
    $saved_fids = $saveData['save_as']['save_as_top_fids'];
    
    // 步骤5: 清理广告文件（可选）
    if (ENABLE_AD_FILTER && !empty(QUARK_BANNED_KEYWORDS)) {
        $saved_fids = filterAdFiles($saved_fids, QUARK_BANNED_KEYWORDS);
        if (empty($saved_fids)) {
            throw new Exception('资源内容为空（全部被过滤）');
        }
    }
    
    // 步骤6: 分享转存后的资源
    if (!empty($ad_fid)) {
        $saved_fids[] = $ad_fid;
    }
    $shareRes = createShare($saved_fids, $title, $expired_type);
    if ($shareRes['status'] !== 200) {
        throw new Exception($shareRes['message'] ?? '创建分享失败');
    }
    $share_task_id = $shareRes['data']['task_id'];
    
    // 步骤7: 轮询分享任务获取share_id
    $shareData = pollTask($share_task_id);
    if (!$shareData || $shareData['status'] != 2) {
        throw new Exception('分享任务超时或失败');
    }
    $share_id = $shareData['share_id'];
    
    // 步骤8: 获取分享链接和提取码
    $passwordRes = getSharePassword($share_id);
    if ($passwordRes['status'] !== 200) {
        throw new Exception($passwordRes['message'] ?? '获取分享链接失败');
    }
    
    $shareInfo = $passwordRes['data'];
    
    return [
        'share_url' => $shareInfo['share_url'] ?? '',
        'code' => $shareInfo['passcode'] ?? '',
        'title' => $title,
        'fid' => (is_array($saved_fids) && count($saved_fids) > 1) 
                 ? $saved_fids 
                 : ($shareInfo['first_file']['fid'] ?? ''),
    ];
}

/**
 * 获取分享令牌
 */
function getStoken($pwd_id, $passcode) {
    $data = [
        'passcode' => $passcode,
        'pwd_id' => $pwd_id,
    ];
    $queryParams = [
        'pr' => 'ucpro',
        'fr' => 'pc',
        'uc_param_str' => '',
    ];
    return apiRequest(
        'https://drive-pc.quark.cn/1/clouddrive/share/sharepage/token',
        'POST',
        $data,
        $queryParams
    );
}

/**
 * 获取分享资源详情
 */
function getShareDetail($pwd_id, $stoken) {
    $queryParams = [
        'pr' => 'ucpro',
        'fr' => 'pc',
        'uc_param_str' => '',
        'pwd_id' => $pwd_id,
        'stoken' => $stoken,
        'pdir_fid' => '0',
        'force' => '0',
        '_page' => '1',
        '_size' => '100',
        '_fetch_banner' => '1',
        '_fetch_share' => '1',
        '_fetch_total' => '1',
        '_sort' => 'file_type:asc,updated_at:desc'
    ];
    return apiRequest(
        'https://drive-pc.quark.cn/1/clouddrive/share/sharepage/detail',
        'GET',
        [],
        $queryParams
    );
}

/**
 * 转存资源到自己网盘
 */
function saveShareToMyDrive($pwd_id, $stoken, $fid_list, $fid_token_list, $to_pdir_fid) {
    $data = [
        'fid_list' => $fid_list,
        'fid_token_list' => $fid_token_list,
        'to_pdir_fid' => $to_pdir_fid,
        'pwd_id' => $pwd_id,
        'stoken' => $stoken,
        'pdir_fid' => '0',
        'scene' => 'link',
    ];
    $queryParams = [
        'entry' => 'update_share',
        'pr' => 'ucpro',
        'fr' => 'pc',
        'uc_param_str' => ''
    ];
    return apiRequest(
        'https://drive-pc.quark.cn/1/clouddrive/share/sharepage/save',
        'POST',
        $data,
        $queryParams
    );
}

/**
 * 创建分享
 */
function createShare($fid_list, $title, $expired_type) {
    $data = [
        'fid_list' => $fid_list,
        'expired_type' => $expired_type,
        'title' => $title,
        'url_type' => 1,
    ];
    $queryParams = [
        'pr' => 'ucpro',
        'fr' => 'pc',
        'uc_param_str' => ''
    ];
    return apiRequest(
        'https://drive-pc.quark.cn/1/clouddrive/share',
        'POST',
        $data,
        $queryParams
    );
}

/**
 * 获取分享密码和链接
 */
function getSharePassword($share_id) {
    $data = [
        'share_id' => $share_id,
    ];
    $queryParams = [
        'pr' => 'ucpro',
        'fr' => 'pc',
        'uc_param_str' => ''
    ];
    return apiRequest(
        'https://drive-pc.quark.cn/1/clouddrive/share/password',
        'POST',
        $data,
        $queryParams
    );
}

/**
 * 轮询任务状态
 */
function pollTask($task_id) {
    $retry_index = 0;
    $taskData = null;
    
    while ($retry_index < MAX_RETRY_COUNT) {
        $queryParams = [
            'pr' => 'ucpro',
            'fr' => 'pc',
            'uc_param_str' => '',
            'task_id' => $task_id,
            'retry_index' => $retry_index
        ];
        
        $res = apiRequest(
            'https://drive-pc.quark.cn/1/clouddrive/task',
            'GET',
            [],
            $queryParams
        );
        
        if ($res['status'] === 200) {
            $taskData = $res['data'];
            if ($taskData['status'] == 2) {
                return $taskData;
            }
        }
        
        // 检查特殊错误
        if (isset($res['message']) && strpos($res['message'], 'capacity limit') !== false) {
            throw new Exception('网盘容量不足');
        }
        
        $retry_index++;
        usleep(500000); // 等待0.5秒
    }
    
    return $taskData;
}

/**
 * 过滤广告文件（递归扫描所有子文件夹）
 */
function filterAdFiles($fid_list, $banned_keywords) {
    global $filterLog;
    $filterLog = [
        'scan_start_time' => date('Y-m-d H:i:s'),
        'banned_keywords' => [],
        'structure' => [],
        'total_files' => 0,
        'total_folders' => 0,
        'ad_files' => [],
        'deleted_fids' => [],
        'result' => 'success'
    ];
    
    if (empty($fid_list)) {
        $filterLog['result'] = 'empty_fid_list';
        return $fid_list;
    }
    
    $bannedList = array_map('trim', explode(',', $banned_keywords));
    $pdir_fid = is_array($fid_list) ? $fid_list[0] : $fid_list;
    
    // 输出广告关键词黑名单配置
    $filterLog['banned_keywords'] = $bannedList;
    // 只过滤空字符串，保留"0"等有效关键词
    $validKeywords = array_filter($bannedList, function($k) { return $k !== ''; });
    $keywordDisplay = empty($validKeywords) ? '未配置' : implode(', ', $validKeywords);
    $filterLog['structure'][] = "🚫 广告关键词黑名单: [{$keywordDisplay}]";
    
    // 🆕 递归获取所有文件和文件夹
    $filterLog['structure'][] = "📂 开始递归扫描根目录: fid={$pdir_fid}";
    $allItems = getAllItemsRecursively($pdir_fid, 0);
    
    if (empty($allItems)) {
        $filterLog['result'] = 'empty_file_list';
        $filterLog['structure'][] = "⚠️ 未找到任何文件或文件夹";
        return $fid_list;
    }
    
    $filterLog['total_files'] = count(array_filter($allItems, function($item) {
        return ($item['dir'] ?? '') != 1;
    }));
    $filterLog['total_folders'] = count(array_filter($allItems, function($item) {
        return ($item['dir'] ?? '') == 1;
    }));
    
    $filterLog['structure'][] = "📊 扫描完成: 共 {$filterLog['total_files']} 个文件, {$filterLog['total_folders']} 个文件夹";
    
    // 检查广告关键词
    $delList = [];
    foreach ($allItems as $item) {
        $fileName = $item['file_name'] ?? '';
        $fileType = ($item['dir'] ?? '') == 1 ? '📁' : '📄';
        $depth = $item['depth'] ?? 0;
        $indent = str_repeat('  ', $depth);
        
        $matchedKeyword = null;
        foreach ($bannedList as $keyword) {
            // 使用 !== '' 而不是 !empty()，避免 "0" 被误判为空
            if ($keyword !== '' && strpos($fileName, $keyword) !== false) {
                $matchedKeyword = $keyword;
                break;
            }
        }
        
        if ($matchedKeyword) {
            $delList[] = $item['fid'];
            $filterLog['ad_files'][] = [
                'name' => $fileName,
                'fid' => $item['fid'],
                'type' => ($item['dir'] ?? '') == 1 ? 'folder' : 'file',
                'keyword' => $matchedKeyword,
                'path' => $item['path'] ?? ''
            ];
            $filterLog['structure'][] = "{$indent}{$fileType} {$fileName} ❌ [匹配关键词: {$matchedKeyword}]";
        } else {
            $filterLog['structure'][] = "{$indent}{$fileType} {$fileName} ✅";
        }
    }
    
    // 如果全部是广告，删除根目录并返回空
    if (count($delList) === count($allItems)) {
        $filterLog['result'] = 'all_files_are_ads';
        $filterLog['structure'][] = "🚫 所有文件都包含广告关键词，删除根目录";
        deleteFiles(is_array($fid_list) ? $fid_list : [$fid_list]);
        $filterLog['deleted_fids'] = is_array($fid_list) ? $fid_list : [$fid_list];
        return [];
    }
    
    // 删除广告文件
    if (!empty($delList)) {
        $filterLog['structure'][] = "🗑️ 开始删除 " . count($delList) . " 个广告文件/文件夹";
        deleteFiles($delList);
        $filterLog['deleted_fids'] = $delList;
        $filterLog['result'] = 'partial_deletion';
    } else {
        $filterLog['structure'][] = "✅ 未检测到广告文件";
        $filterLog['result'] = 'no_ads_found';
    }
    
    $filterLog['scan_end_time'] = date('Y-m-d H:i:s');
    
    return $fid_list;
}

/**
 * 递归获取文件夹下所有文件和文件夹（包括子文件夹）
 * 用于广告词检测，同时扫描文件名和文件夹名
 * 
 * @param string $pdir_fid 文件夹ID
 * @param int $depth 当前递归深度
 * @param string $parentPath 父路径
 * @return array 所有文件和文件夹列表
 */
function getAllItemsRecursively($pdir_fid, $depth = 0, $parentPath = '') {
    global $filterLog;
    $allItems = [];
    
    try {
        // 获取当前文件夹的直接子项
        $currentList = getFileList($pdir_fid);
        
        if (empty($currentList)) {
            return [];
        }
        
        foreach ($currentList as $item) {
            // 构建当前路径
            $currentPath = $parentPath ? ($parentPath . '/' . $item['file_name']) : $item['file_name'];
            
            // 添加深度和路径信息
            $item['depth'] = $depth;
            $item['path'] = $currentPath;
            
            // 将当前项加入结果（无论是文件还是文件夹都要检查）
            $allItems[] = $item;
            
            // 如果是文件夹，递归获取其中的内容
            if (isset($item['dir']) && $item['dir'] == 1) {
                if (isset($filterLog)) {
                    $indent = str_repeat('  ', $depth);
                    $filterLog['structure'][] = "{$indent}  ↳ 递归进入文件夹: {$item['file_name']}";
                }
                $subItems = getAllItemsRecursively($item['fid'], $depth + 1, $currentPath);
                if (!empty($subItems)) {
                    $allItems = array_merge($allItems, $subItems);
                }
            }
        }
    } catch (Exception $e) {
        // 递归过程中出错，记录但不中断
        if (isset($filterLog)) {
            $filterLog['structure'][] = "⚠️ 递归扫描出错 (depth={$depth}): " . $e->getMessage();
        }
    }
    
    return $allItems;
}

/**
 * 获取文件列表
 */
function getFileList($pdir_fid) {
    $queryParams = [
        'pr' => 'ucpro',
        'fr' => 'pc',
        'uc_param_str' => '',
        'pdir_fid' => $pdir_fid,
        '_page' => 1,
        '_size' => 200,
        '_fetch_total' => 1,
        '_fetch_sub_dirs' => 1,
        '_sort' => 'file_type:asc,updated_at:desc',
    ];
    
    $res = apiRequest(
        'https://drive-pc.quark.cn/1/clouddrive/file/sort',
        'GET',
        [],
        $queryParams
    );
    
    if ($res['status'] === 200 && isset($res['data']['list'])) {
        return $res['data']['list'];
    }
    
    return [];
}

/**
 * 删除文件
 */
function deleteFiles($fid_list) {
    $data = [
        'action_type' => 2,
        'exclude_fids' => [],
        'filelist' => $fid_list,
    ];
    $queryParams = [
        'pr' => 'ucpro',
        'fr' => 'pc',
        'uc_param_str' => ''
    ];
    return apiRequest(
        'https://drive-pc.quark.cn/1/clouddrive/file/delete',
        'POST',
        $data,
        $queryParams
    );
}

/**
 * API请求封装
 */
function apiRequest($url, $method, $data, $queryParams = []) {
    // 构建查询参数
    if (!empty($queryParams)) {
        $url .= '?' . http_build_query($queryParams);
    }
    
    // 构建请求头
    $headers = [
        'Accept: application/json, text/plain, */*',
        'Accept-Language: zh-CN,zh;q=0.9',
        'Content-Type: application/json;charset=UTF-8',
        'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Referer: https://pan.quark.cn/',
        'Cookie: ' . QUARK_COOKIE
    ];
    
    // 初始化cURL
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
    curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_TIMEOUT, API_TIMEOUT);
    curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, API_TIMEOUT);
    
    // 设置请求方法
    if (strtoupper($method) === 'POST') {
        curl_setopt($ch, CURLOPT_POST, true);
        if (!empty($data)) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
        }
    }
    
    // 执行请求
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    
    if ($response === false) {
        $error = curl_error($ch);
        curl_close($ch);
        throw new Exception('请求失败: ' . $error);
    }
    
    curl_close($ch);
    
    // 解析响应
    $result = json_decode($response, true);
    if (json_last_error() !== JSON_ERROR_NONE) {
        throw new Exception('响应解析失败');
    }
    
    return $result;
}

/**
 * 输出JSON响应
 */
function outputJson($code, $message, $data = null) {
    $response = [
        'code' => $code,
        'message' => $message,
        'timestamp' => date('Y-m-d H:i:s')
    ];
    
    if ($data !== null) {
        $response['data'] = $data;
    }
    
    echo json_encode($response, JSON_UNESCAPED_UNICODE | JSON_PRETTY_PRINT);
    exit;
}

