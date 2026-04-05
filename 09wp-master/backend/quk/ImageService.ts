import { injectable } from "inversify";
import axios, { AxiosInstance } from "axios";
import tunnel from "tunnel";
import GlobalSetting from "../models/GlobalSetting";
import { GlobalSettingAttributes } from "../models/GlobalSetting";

@injectable()
export class ImageService {
  private axiosInstance: AxiosInstance | null = null;

  constructor() {
    // 移除构造函数中的初始化，改为懒加载
  }

  private async ensureAxiosInstance(): Promise<AxiosInstance> {
    if (!this.axiosInstance) {
      const settings = await GlobalSetting.findOne();
      const globalSetting = settings?.dataValues || ({} as GlobalSettingAttributes);

      this.axiosInstance = axios.create({
        timeout: 30000,
        headers: {
          Accept: "image/*, */*",
          "User-Agent":
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
        },
        withCredentials: false,
        maxRedirects: 5,
        httpsAgent: globalSetting.isProxyEnabled
          ? tunnel.httpsOverHttp({
              proxy: {
                host: globalSetting.httpProxyHost,
                port: globalSetting.httpProxyPort,
                headers: {
                  "Proxy-Authorization": "",
                },
              },
            })
          : undefined,
      });

      this.axiosInstance.interceptors.response.use(
        (response) => response,
        (error) => {
          throw error;
        }
      );
    }
    return this.axiosInstance;
  }

  async updateAxiosInstance(): Promise<void> {
    this.axiosInstance = null;
    await this.ensureAxiosInstance();
  }

  async getImages(url: string): Promise<any> {
    const axiosInstance = await this.ensureAxiosInstance();

    return await axiosInstance.get(url, {
      responseType: "stream",
      validateStatus: (status) => status >= 200 && status < 300,
      headers: {
        Referer: new URL(url).origin,
      },
    });
  }
}
