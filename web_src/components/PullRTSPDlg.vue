<template>
    <FormDlg title="添加拉流" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any() || bLoading">
        <div :class="['form-group', { 'has-error': errors.has('url')}]">
            <label for="input-url" class="col-sm-3 control-label"><span class="text-red">*</span> RTSP地址</label>
            <div class="col-sm-8">
                <input type="text" id="input-url" class="form-control" name="url" data-vv-as="RTSP地址" v-validate="'required'" v-model.trim="form.source">
                <span class="help-block">{{errors.first('url')}}</span>
            </div>
        </div>                   
        <div :class="['form-group', { 'has-error': errors.has('customPath')}]">
            <label for="input-custom-path" class="col-sm-3 control-label">自定义路径</label>
            <div class="col-sm-8">
                <input type="text" id="input-custom-path" class="form-control" name="customPath" data-vv-as="自定义路径" v-model.trim="form.customPath" placeholder="/your/custom/path">
                <span class="help-block">{{errors.first('customPath')}}</span>
            </div>
        </div>
        <div :class="['form-group', { 'has-error': errors.has('outIp')}]">
            <label for="input-custom-path" class="col-sm-3 control-label">输出ip地址</label>
            <div class="col-sm-8">
                <input type="text" id="input-out-ip" class="form-control" name="customPath" data-vv-as="输出ip地址" v-model.trim="form.outIp" placeholder="/your/out/id">
                <span class="help-block">{{errors.first('outIp')}}</span>
            </div>
        </div>
    </FormDlg>
</template>

<script>
import Vue from 'vue'
import FormDlg from 'components/FormDlg.vue'
import $ from 'jquery'

export default {
    data() {
        return {
            bLoading: false,
            form: this.defForm()
        }
    },
    components: {
        FormDlg
    },
    methods: {
        defForm() {
            return {
                source: '',
                customPath: '',
                outIp: '',
            }
        },
        onHide() {
            this.errors.clear();
            this.form = this.defForm();
        },
        onShow() {
            document.querySelector(`[name=url]`).focus();
        },
        async onSubmit() {
            var ok = await this.$validator.validateAll();
            if(!ok) {
                var e = this.errors.items[0];
                this.$message({
                    type: 'error',
                    message: e.msg
                });
                document.querySelector(`[name=${e.field}]`).focus();
                return;
            }
            this.bLoading = true;
            $.get('/api/v1/stream/add', this.form).then(data => {
                this.$refs['dlg'].hide();
                this.$emit('submit');
            }).always(() => {
                this.bLoading = false;
            })
        },
        show(data) {
            this.errors.clear();
            if(data) {
                Object.assign(this.form, data);
            }
            this.$refs['dlg'].show();
        }
    }
}
</script>
