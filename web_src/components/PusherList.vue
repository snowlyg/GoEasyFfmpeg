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
                <form class="form-inline">
                    <div class="form-group">
                        <button type="button" class="btn btn-sm btn-success"
                                @click.prevent="$refs['pullRTSPDlg'].show()"><i class="fa fa-plus"></i> 拉流分发
                        </button>
                        <button type="button" class="btn btn-sm btn-warning" @click.prevent="startAll()">启动</button>
                        <button type="button" class="btn btn-sm btn-danger" @click="stopAll()">停止</button>
                    </div>
                    <div class="form-group pull-right">
                        <div class="input-group">
                            <input type="text" class="form-control" placeholder="搜索" v-model.trim="q"
                                   @keydown.enter.prevent ref="q">
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-default" @click.prevent="doSearch">
                                    <i class="fa fa-search"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                </form>
            </div>
            <div class="box-body">
                <el-table :data="pushers" stripe class="view-list" :default-sort="{prop: 'Id', order: 'descending'}"
                          @sort-change="sortChange" @selection-change="handleSelectionChange">
                    <el-table-column
                            type="selection"
                            width="55">
                    </el-table-column>
                    <el-table-column prop="id" label="ID" min-width="60"></el-table-column>
                    <el-table-column label="播放地址" min-width="240" show-overflow-tooltip>
                        <template slot-scope="scope">
                        <span>
                          <i class="fa fa-copy" role="button" v-clipboard="scope.row.url" title="点击拷贝"
                             @success="$message({type:'success', message:'成功拷贝到粘贴板'})"></i>
                          {{scope.row.url}}
                          </span>
                        </template>
                    </el-table-column>
                    <el-table-column label="源地址" min-width="240" show-overflow-tooltip>
                        <template slot-scope="scope">
                        <span v-if="scope.row.source">
                          <i class="fa fa-copy" role="button" v-clipboard="scope.row.source" title="点击拷贝"
                             @success="$message({type:'success', message:'成功拷贝到粘贴板'})"></i>
                          {{scope.row.source}}
                          </span>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                    <el-table-column prop="transType" label="传输方式" min-width="100"></el-table-column>
                    <el-table-column prop="startAt" label="开始时间" min-width="200" sortable="custom"></el-table-column>
                    <el-table-column prop="status" label="状态" min-width="70" sortable="custom">
                        <template slot-scope="scope">
                            <span v-if="scope.row.status=== '已启动'" class="bg-green" style="padding: 3px">{{scope.row.status}}</span>
                            <span v-else class="bg-red" style="padding: 3px">{{scope.row.status}}</span>
                        </template>
                    </el-table-column>
                    <el-table-column label="操作" min-width="200" fixed="right">
                        <template slot-scope="scope">
                            <div class="btn-group">
                                <span v-if="scope.row.status === '已启动'">
                                     <a v-if="scope.row.status === '已启动'" role="button" class="btn btn-xs btn-warning"
                                        @click.prevent="stop(scope.row)">
                                  <i class="fa fa-stop"></i> 停止
                                </a>
                                </span>
                                <span v-else>
                                     <a role="button" class="btn btn-xs btn-success" @click.prevent="edit(scope.row)">
                                      <i class="fa fa-edit"></i> 编辑
                                    </a>
                                    <a role="button" class="btn btn-xs btn-success" @click.prevent="start(scope.row)">
                                      <i class="fa fa-play"></i> 启动
                                    </a>
                                    <a v-if="scope.row.status === '已停止'" role="button" class="btn btn-xs btn-danger"
                                       @click.prevent="del(scope.row)"><i class="fa fa-trash"></i> 删除
                                    </a>
                                </span>
                            </div>
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
            // this.timer = setInterval(() => {
            this.getPushers();
            // }, 3000);
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
            stop(row) {
                this.$confirm(`确认停止 ${row.id} ?`, "提示").then(() => {
                    $.get("/api/v1/stream/stop", {
                        id: row.id
                    }).then(data => {
                        this.getPushers();
                    })
                }).catch(() => {
                });
            },
            start(row) {
                this.$confirm(`确认启动 ${row.id} ?`, "提示").then(() => {
                    $.get("/api/v1/stream/start", {
                        id: row.id
                    }).then(data => {
                        this.getPushers();
                    })
                }).catch(() => {
                });
            },
            del(row) {
                this.$confirm(`确认删除 ${row.id} ?`, "提示").then(() => {
                    $.get("/api/v1/stream/del", {
                        id: row.id
                    }).then(data => {
                        this.getPushers();
                    })
                }).catch(() => {
                });
            },
            edit(row) {
                this.$refs['editRTSPDlg'].show(row)
            },
            stopAll() {
                for (let i = 0; i < this.multipleSelection.length; i++) {
                    this.ids += this.multipleSelection[i].id+",";
                }

                $.post("/api/v1/stream/stopAll", {
                    ids:  JSON.stringify(this.ids)
                }).then(data => {
                    this.getPushers();
                })
            },
            startAll() {
                for (let i = 0; i < this.multipleSelection.length; i++) {
                    this.ids += this.multipleSelection[i].id + ",";
                }
                $.post("/api/v1/stream/startAll", {
                    ids: JSON.stringify( this.ids)
                }).then(data => {
                    this.getPushers();
                })
            },
            handleSelectionChange(val) {
                this.multipleSelection = val;
            }
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
    };
</script>

