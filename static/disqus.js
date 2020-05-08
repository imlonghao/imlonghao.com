if (document.getElementById("disqus_thread") !== null) {
  var dsqjs = new DisqusJS({
    shortname: "imlonghao",
    siteName: "imlonghao",
    identifier: document.location.origin + document.location.pathname,
    url: document.location.origin + document.location.pathname,
    api: "https://disqus.skk.moe/disqus/",
    apikey: "AJMeKW6aKyA7j6bMPl9MlSsVczcxKZIhpCqi8HBdVpZ2oY9utXLqSFysRA0FLxlF",
    admin: "imlonghao",
    adminLabel: "Mod",
  });
}
