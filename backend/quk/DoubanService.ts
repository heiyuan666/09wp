import { AxiosHeaders, AxiosInstance } from "axios";
import { createAxiosInstance } from "../utils/axiosInstance";

interface DoubanSubject {
  id: string;
  title: string;
  rate: string;
  cover: string;
  url: string;
  is_new: boolean;
}

export class DoubanService {
  private baseUrl: string;
  private api: AxiosInstance;

  constructor() {
    this.baseUrl = "https://movie.douban.com/j";
    this.api = createAxiosInstance(
      this.baseUrl,
      AxiosHeaders.from({
        accept: "*/*",
        "accept-language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
        priority: "u=1, i",
        "sec-ch-ua": '"Not A(Brand";v="8", "Chromium";v="132", "Microsoft Edge";v="132"',
        "sec-ch-ua-mobile": "?0",
        "sec-ch-ua-platform": '"Windows"',
        "sec-fetch-dest": "empty",
        "sec-fetch-mode": "cors",
        "sec-fetch-site": "same-origin",
        "x-requested-with": "XMLHttpRequest",
        cookie:
          'll="118282"; bid=StA6AQFsAWQ; _pk_id.100001.4cf6=6448be57b1b5ca7e.1723172321.; _vwo_uuid_v2=DC15B8183560FF1E538FFE1D480723310|c08e2d213ecb5510005f90a6ff332121; __utmv=30149280.6282; _vwo_uuid_v2=DC15B8183560FF1E538FFE1D480723310|c08e2d213ecb5510005f90a6ff332121; __utmz=30149280.1731915179.21.6.utmcsr=search.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/movie/subject_search; __utmz=223695111.1731915179.21.6.utmcsr=search.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/movie/subject_search; douban-fav-remind=1; __utmc=30149280; __utmc=223695111; _pk_ref.100001.4cf6=%5B%22%22%2C%22%22%2C1739176523%2C%22https%3A%2F%2Fsearch.douban.com%2Fmovie%2Fsubject_search%3Fsearch_text%3D%E8%84%91%E6%B4%9E%E5%A4%A7%E5%BC%80%26cat%3D1002%22%5D; _pk_ses.100001.4cf6=1; ap_v=0,6.0; __utma=30149280.859303574.1723448979.1739167503.1739176523.42; __utmb=30149280.0.10.1739176523; __utma=223695111.1882744177.1723448979.1739167503.1739176523.42; __utmb=223695111.0.10.1739176523',
        Referer: "https://movie.douban.com/",
        "Referrer-Policy": "unsafe-url",
      })
    );
  }

  async getHotList(params: {
    type: string;
    tag: string;
    page_limit: string;
    page_start: string;
  }): Promise<{ data: DoubanSubject[] }> {
    try {
      const response = await this.api.get("/search_subjects", {
        params: params,
      });
      if (response.data && response.data.subjects) {
        return {
          data: response.data.subjects,
        };
      } else {
        return {
          data: [],
        };
      }
    } catch (error) {
      console.error("Error fetching hot list:", error);
      throw error;
    }
  }
}
