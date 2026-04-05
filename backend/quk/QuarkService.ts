import { AxiosInstance, AxiosHeaders } from "axios";
import { logger } from "../utils/logger";
import { createAxiosInstance } from "../utils/axiosInstance";
import { injectable } from "inversify";
import { Request } from "express";
import UserSetting from "../models/UserSetting";
import {
  ShareInfoResponse,
  FolderListResponse,
  QuarkFolderItem,
  SaveFileParams,
} from "../types/cloud";
import { ICloudStorageService } from "@/types/services";

interface QuarkShareInfo {
  stoken?: string;
  pwdId?: string;
  fileSize?: number;
  list: {
    fid: string;
    file_name: string;
    file_type: number;
    share_fid_token: string;
  }[];
}

@injectable()
export class QuarkService implements ICloudStorageService {
  private api: AxiosInstance;
  private cookie: string = "";

  constructor() {
    this.api = createAxiosInstance(
      "https://drive-h.quark.cn",
      AxiosHeaders.from({
        cookie: this.cookie,
        accept: "application/json, text/plain, */*",
        "accept-language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
        "content-type": "application/json",
        priority: "u=1, i",
        "sec-ch-ua": '"Microsoft Edge";v="131", "Chromium";v="131", "Not_A Brand";v="24"',
        "sec-ch-ua-mobile": "?0",
        "sec-ch-ua-platform": '"Windows"',
        "sec-fetch-dest": "empty",
        "sec-fetch-mode": "cors",
        "sec-fetch-site": "same-site",
      })
    );

    this.api.interceptors.request.use((config) => {
      config.headers.cookie = this.cookie;
      return config;
    });
  }

  async setCookie(req: Request): Promise<void> {
    const userId = req.user?.userId;
    const userSetting = await UserSetting.findOne({
      where: { userId },
    });
    if (userSetting && userSetting.dataValues.quarkCookie) {
      this.cookie = userSetting.dataValues.quarkCookie;
    } else {
      throw new Error("请先设置夸克网盘cookie");
    }
  }

  async getShareInfo(pwdId: string, passcode = ""): Promise<ShareInfoResponse> {
    const response = await this.api.post(
      `/1/clouddrive/share/sharepage/token?pr=ucpro&fr=pc&uc_param_str=&__dt=994&__t=${Date.now()}`,
      {
        pwd_id: pwdId,
        passcode,
      }
    );
    if (response.data?.status === 200 && response.data.data) {
      const fileInfo = response.data.data;
      if (fileInfo.stoken) {
        const res = await this.getShareList(pwdId, fileInfo.stoken);
        return {
          data: res,
        };
      }
    }
    throw new Error("获取夸克分享信息失败");
  }

  async getShareList(pwdId: string, stoken: string): Promise<ShareInfoResponse["data"]> {
    const response = await this.api.get("/1/clouddrive/share/sharepage/detail", {
      params: {
        pr: "ucpro",
        fr: "pc",
        uc_param_str: "",
        pwd_id: pwdId,
        stoken: stoken,
        pdir_fid: "0",
        force: "0",
        _page: "1",
        _size: "50",
        _fetch_banner: "1",
        _fetch_share: "1",
        _fetch_total: "1",
        _sort: "file_type:asc,updated_at:desc",
        __dt: "1589",
        __t: Date.now(),
      },
    });
    if (response.data?.data) {
      const list = response.data.data.list
        .filter((item: QuarkShareInfo["list"][0]) => item.fid)
        .map((folder: QuarkShareInfo["list"][0]) => ({
          fileId: folder.fid,
          fileName: folder.file_name,
          fileIdToken: folder.share_fid_token,
        }));
      return {
        list,
        pwdId,
        stoken,
        fileSize: response.data.data.share?.size || 0,
      };
    } else {
      return {
        list: [],
      };
    }
  }

  async getFolderList(parentCid = "0"): Promise<FolderListResponse> {
    const response = await this.api.get("/1/clouddrive/file/sort", {
      params: {
        pr: "ucpro",
        fr: "pc",
        uc_param_str: "",
        pdir_fid: parentCid,
        _page: "1",
        _size: "100",
        _fetch_total: "false",
        _fetch_sub_dirs: "1",
        _sort: "",
        __dt: "2093126",
        __t: Date.now(),
      },
    });
    if (response.data?.data && response.data.data.list) {
      const data = response.data.data.list
        .filter((item: QuarkFolderItem) => item.fid && item.file_type === 0)
        .map((folder: QuarkFolderItem) => ({
          cid: folder.fid,
          name: folder.file_name,
          path: [],
        }));
      return {
        data,
      };
    } else {
      const message = "获取夸克目录列表失败:" + response.data.error;
      logger.error(message);
      throw new Error(message);
    }
  }

  async saveSharedFile(params: SaveFileParams): Promise<{ message: string; data: unknown }> {
    const quarkParams = {
      fid_list: params.fids,
      fid_token_list: params.fidTokens,
      to_pdir_fid: params.folderId,
      pwd_id: params.shareCode,
      stoken: params.receiveCode,
      pdir_fid: "0",
      scene: "link",
    };
    try {
      const response = await this.api.post(
        `/1/clouddrive/share/sharepage/save?pr=ucpro&fr=pc&uc_param_str=&__dt=208097&__t=${Date.now()}`,
        quarkParams
      );

      return {
        message: response.data.message,
        data: response.data.data,
      };
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : "未知错误");
    }
  }
}
