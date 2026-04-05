import axios from 'axios';

// 创建统一的 axios 实例
export const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
});

// 请求拦截器：添加管理员密码到请求头
api.interceptors.request.use(
  (config) => {
    // 从 localStorage 读取密码
    const password = localStorage.getItem('admin_password');
    if (password) {
      config.headers['X-Admin-Password'] = password;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器：处理 401 未授权错误
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // 清除认证状态
      localStorage.removeItem('admin_authenticated');
      localStorage.removeItem('admin_password');
      // 跳转到登录页
      if (window.location.pathname !== '/admin/login') {
        window.location.href = '/admin/login';
      }
    }
    return Promise.reject(error);
  }
);

export interface LoginRequest {
  password: string;
}

export interface LoginResponse {
  success: boolean;
  message: string;
}

export const authApi = {
  // 管理员登录
  login: async (password: string): Promise<LoginResponse> => {
    // 登录接口不需要密码头，因为还没有认证
    const response = await axios.post<LoginResponse>('/api/v1/auth/login', { password });
    return response.data;
  },
};



