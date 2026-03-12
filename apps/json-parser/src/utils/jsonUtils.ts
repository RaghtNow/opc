/**
 * JSON 解析工具 - 核心功能模块
 * @version 1.0.1
 * @date 2026-03-12
 * CI/CD 验证标记：部署流程测试
 */

/**
 * 解析 JSON 字符串
 */
export function parseJson(input: string): { 
  success: boolean; 
  data?: any; 
  error?: string;
  errorLine?: number;
  errorColumn?: number;
} {
  try {
    const data = JSON.parse(input);
    return { success: true, data };
  } catch (e: any) {
    // 尝试提取错误位置
    const match = e.message.match(/position (\d+)/);
    let errorLine = 1;
    let errorColumn = 1;
    
    if (match) {
      const position = parseInt(match[1]);
      const lines = input.substring(0, position).split('\n');
      errorLine = lines.length;
      errorColumn = lines[lines.length - 1].length + 1;
    }
    
    return { 
      success: false, 
      error: e.message,
      errorLine,
      errorColumn
    };
  }
}

/**
 * 格式化 JSON（美化输出）
 */
export function formatJson(data: any, indent: number = 2): string {
  return JSON.stringify(data, null, indent);
}

/**
 * 压缩 JSON（移除空白）
 */
export function minifyJson(data: any): string {
  return JSON.stringify(data);
}

/**
 * 验证 JSON 并返回详细信息
 */
export function validateJson(input: string): {
  isValid: boolean;
  type?: string;
  size?: number;
  keys?: string[];
  error?: string;
  errorLine?: number;
  errorColumn?: number;
} {
  const result = parseJson(input);
  
  if (!result.success) {
    return {
      isValid: false,
      error: result.error,
      errorLine: result.errorLine,
      errorColumn: result.errorColumn
    };
  }
  
  const data = result.data;
  const type = getType(data);
  
  let info: any = {
    isValid: true,
    type,
    size: input.length
  };
  
  // 如果是对象或数组，提供键列表
  if (type === 'object') {
    info.keys = Object.keys(data);
    info.keyCount = info.keys.length;
  } else if (type === 'array') {
    info.length = data.length;
  }
  
  return info;
}

/**
 * 获取数据的类型
 */
function getType(data: any): string {
  if (data === null) return 'null';
  if (Array.isArray(data)) return 'array';
  return typeof data;
}

/**
 * 检查 JSON 是否为有效结构
 */
export function isValidJsonStructure(input: string): boolean {
  try {
    JSON.parse(input);
    return true;
  } catch {
    return false;
  }
}
