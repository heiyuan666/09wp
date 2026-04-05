const links = [
  '\u641c\u7d22\u5f15\u64ce',
  '\u672c\u7ad9\u670d\u52a1\u5668\u8d2d\u4e70',
  '\u963f\u91cc\u76d8\u641c',
  '\u5938\u514b\u641c',
  '\u5168\u80fd\u4e91\u76d8\u64ad\u653e\u5668',
];

const disclaimer =
  '\u656c\u544a\u4e0e\u58f0\u660e\uff1a\u672c\u7ad9\u4e0d\u4ea7\u751f/\u5b58\u50a8\u4efb\u4f55\u6570\u636e\uff0c\u4e5f\u4ece\u672a\u53c2\u4e0e\u5f55\u5236\u3001\u4e0a\u4f20\uff0c\u6240\u6709\u8d44\u6e90\u5747\u6765\u81ea\u7f51\u7edc\uff0c\u53ca\u7f51\u53cb\u63d0\u4ea4\uff1b\u65e0\u610f\u5192\u72af\u4efb\u4f55\u516c\u53f8\u3001\u7528\u6237\u7684\u6743\u76ca\u3001\u7248\u6743\uff0c\u4f60\u53ef\u4ee5\u901a\u8fc7\u9996\u9875\u9876\u90e8\u7684\u3010\u4fb5\u6743\u5c4f\u853d\u3011\u8fdb\u884c\u5173\u952e\u8bcd\u5c4f\u853d\u3002';

const friendLabel = '\u53cb\u60c5\u94fe\u63a5\uff1a';

export function Footer() {
  return (
    <footer className="border-t border-[#e7e7e7] bg-[#f6f6f6] py-26px">
      <div className="mx-auto max-w-1180px px-20px text-center text-14px leading-1.8 text-[#4b5563]">
        <p className="m-0">{disclaimer}</p>
        <div className="mt-10px flex flex-wrap items-center justify-center gap-10px">
          <span>{friendLabel}</span>
          {links.map((item) => (
            <a key={item} href="#" className="text-[#2563eb] no-underline hover:underline">
              {item}
            </a>
          ))}
        </div>
      </div>
    </footer>
  );
}
