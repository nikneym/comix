(async () => {
  const res = await fetch(
    "https://batcave.biz/engine/ajax/controller.php?mod=api&action=reader%2FgetChapterData",
    {
      credentials: "omit",
      headers: {
        "User-Agent":
          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:157.0) Gecko/20100101 Firefox/157.0",
        Accept: "*/*",
        "Accept-Language": "en-US,en;q=0.9",
        "Content-Type": "application/x-www-form-urlencoded",
        "Upgrade-Insecure-Requests": "1",
        "Sec-Fetch-Dest": "document",
        "Sec-Fetch-Mode": "navigate",
        "Sec-Fetch-Site": "cross-site",
        "Sec-Fetch-User": "?1",
        Priority: "u=4",
        Pragma: "no-cache",
        "Cache-Control": "no-cache",
      },
      //body: "news_id=4908&chapter_id=25706",
      body: "news_id=33086&chapter_id=234290",
      method: "POST",
      mode: "cors",
    },
  );
  return res.json();
})();
