import { useState } from 'react';
import { Cloud, Search, Sun } from 'lucide-react';

const topLinks = [
  '\u805a\u5408\u641c\u7d22',
  '\u2b06\u63d0\u4ea4\u8d44\u6e90',
  '\u4fb5\u6743\u5c4f\u853d',
  '\u6dd8\u5b9d\u9690\u85cf\u4f18\u60e0\u5238',
  '\u672c\u7ad9\u642d\u5efa',
];
const actionCards = [
  '\u672c\u7ad9\u670d\u52a1\u5668',
  '\u6dd8\u5b9d\u9690\u85cf\u4f18\u60e0\u5238',
  '(0905)\u672c\u7ad9APP\u4e0b\u8f7d',
  '\u4e34\u65f6\u57df\u540d(\u53ef\u4fdd\u5b58\u4e66\u7b7e\uff0c\u6bd4\u8f83\u5feb)',
];
const tags = ['\u76f4\u64ad', '\u7231\u60c5', '\u5e05\u54e5', '\u7f8e\u5973', '\u7535\u5f71'];

const siteName = '\u61d2\u76d8\u641c\u7d22';
const loginText = '\u767b\u5f55';
const descText =
  '\u81f4\u529b\u4e8e\u514d\u8d39\u63d0\u4f9b\u5938\u514b\u7f51\u76d8\u3001\u963f\u91cc\u4e91\u76d8\u3001\u8fc5\u96f7\u7f51\u76d8\u7684\u8d44\u6e90';
const searchText = '\u641c\u7d22';
const serviceText = '\u670d\u52a1';
const placeholderText = '\u8f93\u5165\u5173\u952e\u8bcd\u8fdb\u884c\u641c\u7d22';

export function Home() {
  const [keyword, setKeyword] = useState('');

  const handleSearch = () => {
    if (!keyword.trim()) return;
    window.open(`/search?q=${encodeURIComponent(keyword.trim())}`, '_blank');
  };

  return (
    <div className="flex-1 bg-[#f6f6f6] text-[#111827]">
      <header className="h-68px border-b border-[#e7e7e7] bg-[#f7f7f7]">
        <div className="mx-auto h-full max-w-1180px px-20px flex items-center justify-between">
          <div className="flex items-center gap-34px">
            <a href="/" className="flex items-center gap-10px no-underline text-[#111827]">
              <Cloud size={26} className="text-[#7a6a5e]" />
              <span className="text-38px font-700 leading-none">{siteName}</span>
            </a>
            <nav className="hidden md:flex items-center gap-28px text-16px text-[#2b2b2b]">
              {topLinks.map((item) => (
                <a key={item} href="#" className="no-underline text-inherit hover:text-[#165dff]">
                  {item}
                </a>
              ))}
            </nav>
          </div>

          <div className="flex items-center gap-28px text-16px">
            <button
              type="button"
              className="h-34px w-34px rounded-full border-none bg-transparent cursor-pointer text-[#777] hover:bg-[#ececec]"
              aria-label="toggle-theme"
            >
              <Sun size={18} className="m-auto" />
            </button>
            <a href="/admin/login" className="no-underline text-[#222] hover:text-[#165dff]">
              {loginText}
            </a>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-900px px-20px pt-145px pb-120px text-center">
        <div className="flex items-center justify-center gap-12px">
          <Cloud size={44} className="text-[#7a6a5e]" />
          <h1 className="m-0 text-64px leading-none font-700 text-[#0f172a]">{siteName}</h1>
        </div>

        <p className="mx-auto mt-30px mb-0 max-w-760px text-24px leading-1.6 text-[#243042]">
          {descText}
          <a href="#" className="ml-4px text-[#2468ff] no-underline hover:underline">
            {searchText}
          </a>
          {serviceText}
        </p>

        <div className="mx-auto mt-42px max-w-760px flex items-center rounded-full bg-white shadow-[0_8px_26px_rgba(0,0,0,0.08)] border border-[#ededed] px-26px">
          <input
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSearch();
            }}
            placeholder={placeholderText}
            className="h-68px w-full border-none bg-transparent text-30px outline-none placeholder:text-[#9aa0a6]"
          />
          <button
            type="button"
            onClick={handleSearch}
            className="h-42px w-42px flex items-center justify-center rounded-full border-none bg-transparent cursor-pointer text-[#555] hover:bg-[#f3f4f6]"
            aria-label="search"
          >
            <Search size={24} />
          </button>
        </div>

        <div className="mx-auto mt-28px grid max-w-740px grid-cols-1 gap-16px md:grid-cols-2">
          {actionCards.map((item) => (
            <button
              key={item}
              type="button"
              className="h-62px border-none rounded-12px bg-[#efefef] text-27px text-[#303030] cursor-pointer transition hover:bg-[#e7e7e7]"
            >
              {item}
            </button>
          ))}
        </div>

        <div className="mt-34px flex items-center justify-center gap-14px flex-wrap">
          {tags.map((item) => (
            <button
              key={item}
              type="button"
              className="h-44px min-w-84px rounded-10px border-none bg-[#f1f1f1] px-16px text-24px text-[#303030] cursor-pointer hover:bg-[#e9e9e9]"
            >
              {item}
            </button>
          ))}
        </div>
      </main>
    </div>
  );
}
