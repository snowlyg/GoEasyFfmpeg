<template>
  <div class="container-fluid no-padding">
    <div class="alert alert-success alert-dismissible">
      <small>
        <strong><i class="fa fa-info-circle"></i> 提示 : </strong>
        屏幕直播工具可以采用<a href="https://github.com/EasyDSS/EasyScreenLive" target="_blank"> EasyScreenLive <i
          class="fa fa-external-link"></i></a>，
        <span class="push-url-format">推流URL规则: rtsp://{ip}:{port}/{id}</span> ，
        例如 : rtsp://www.easydarwin.org:554/your_stream_id
      </small>
      <button type="button" class="close" data-dismiss="alert" aria-label="Close"><span
          aria-hidden="true">&times;</span></button>
    </div>

    <div class="box box-success">
      <div class="box-header">
        <h4 class="text-success text-center">推流列表</h4>
      </div>
      <div class="box-body">
        <el-table :data="pushers" stripe class="view-list" :default-sort="{prop: 'Id', order: 'descending'}"
                  @sort-change="sortChange" @selection-change="handleSelectionChange">

          <el-table-column label="播放地址" min-width="240" show-overflow-tooltip>
            <template slot-scope="scope">
              <video id=" {{scope.row.Id}}" ref="hlsVideo" controls preload="true"/>
            </template>
          </el-table-column>

        </el-table>
      </div>
      <div class="box-footer clearfix" v-if="total > 0">
        <el-pagination layout="prev,pager,next" class="pull-right" :total="total" :page-size.sync="pageSize"
                       :current-page.sync="currentPage"></el-pagination>
      </div>
    </div>
    <PullRTSPDlg ref="pullRTSPDlg" @submit="getPushers"></PullRTSPDlg>
    <EditRTSPDlg ref="editRTSPDlg" @submit="getPushers" :streamInfo="streamInfo"></EditRTSPDlg>
  </div>
</template>

<script>
import Hls from "hls.js"
import PullRTSPDlg from "components/PullRTSPDlg.vue"
import EditRTSPDlg from "components/EditRTSPDlg.vue"
import prettyBytes from "pretty-bytes";

import _ from "lodash";

export default {
  components: {
    PullRTSPDlg,
    EditRTSPDlg
  },
  props: {},
  data() {
    return {
      q: "",
      sort: "ID",
      order: "descending",
      pushers: [],
      total: 0,
      timer: 0,
      pageSize: 10,
      currentPage: 1,
      streamInfo: {},
      multipleSelection: [],
      ids: "",
    };
  },
  beforeDestroy() {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = 0;
    }
  },
  mounted() {
    this.$refs["q"].focus();
    this.timer = setInterval(() => {
      this.getPushers();
    }, 3000);
  },
  watch: {
    q: function (newVal, oldVal) {
      this.doDelaySearch();
    },
    currentPage: function (newVal, oldVal) {
      this.doSearch(newVal);
    }
  },
  methods: {
    getPushers() {
      $.get("/api/v1/pushers", {
        q: this.q,
        start: (this.currentPage - 1) * this.pageSize,
        limit: this.pageSize,
        sort: this.sort,
        order: this.order
      }).then(data => {
        this.total = data.total;
        this.pushers = data.rows;
        this.pushers.forEach(function (item, index) {
          var video = document.getElementById(item.Id)
          if (Hls.isSupported()) {
            var hls = new Hls()
            hls.loadSource(item.url)
            hls.attachMedia(video)
            hls.on(Hls.Events.MANIFEST_PARSED, function () {
              video.play()
            })
          } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = item.url
            video.addEventListener('loadedmetadata', function () {
              video.play()
            })
          }
        });

      });
    },
    doSearch(page = 1) {
      var query = {};
      if (this.q) query["q"] = this.q;
      this.$router.replace({
        path: `/pushers/${page}`,
        query: query
      });
    },
    doDelaySearch: _.debounce(function () {
      this.doSearch();
    }, 500),
    sortChange(data) {
      this.sort = data.prop;
      this.order = data.order;
      this.getPushers();
    },
    formatBytes(row, col, val) {
      if (val == undefined) return "-";
      return prettyBytes(val);
    },
    beforeRouteEnter(to, from, next) {
      next(vm => {
        vm.q = to.query.q || "";
        vm.currentPage = parseInt(to.params.page) || 1;
      });
    },
    beforeRouteUpdate(to, from, next) {
      next();
      this.$nextTick(() => {
        this.q = to.query.q || "";
        this.currentPage = parseInt(to.params.page) || 1;
        this.pushers = [];
        this.getPushers();
      });
    }
  }
};
</script>

