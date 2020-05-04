(function(factory) {
    if (typeof define === 'function' && define.amd) {
        // AMD. Register as anonymous module.
        define(['jquery'], factory);
    } else if (typeof exports === 'object') {
        // Node / CommonJS
        factory(require('jquery'));
    } else {
        // Browser globals.
        factory(jQuery);
    }
})(function($) {

    'use strict';

    const _ = window._,
        $document = $(document),
        NAMESPACE = 'qor.variations',
        NAME_REPLICATOR = 'qor.replicator',
        EVENT_ENABLE = `enableVariations.${NAMESPACE}`,
        EVENT_DISABLE = `disable.${NAMESPACE}`,
        EVENT_CLICK = `click.${NAMESPACE}`,
        EVENT_KEYUP = `keyup.${NAMESPACE}`,
        EVENT_CHANGED_MEDIALIBRARY = 'changed.medialibrary',
        EVENT_REPLICATOR_ADDED = `addedMultiple.${NAME_REPLICATOR}`,
        EVENT_REPLICATORS_ADDED = `addedMultipleDone.${NAME_REPLICATOR}`,
        CLASS_PRIMARY_META = '.qor-product__property',
        CLASS_SELECT = `${CLASS_PRIMARY_META} select[data-toggle="qor.chooser"]`,
        CLASS_SELECT_TYPE = `${CLASS_PRIMARY_META}-selector`,
        CLASS_TBODY = '.qor-product__table tbody',
        CLASS_TABLE = '.qor-product__table table',
        CLASS_TR = '.qor-product__table tbody tr',
        CLASS_FIELDSET_CONTAINER = '.qor-product__container',
        CLASS_FIELDSET = '.qor-fieldset',
        CLASS_BUTTON_ADD = '.qor-fieldset__add',
        ID_VARIANTS_PRE = 'qor_variants_id_',
        CLASS_SHOULD_REMOVE = 'should_remove',
        CLASS_IS_REMOVE = 'is_removed',
        CLASS_IS_DELETED = '.is_deleted',
        CLASS_TR_SELECTED = `tr.is-selected:not(${CLASS_IS_DELETED})`,
        CLASS_IS_CURRENT = 'is_current',
        CLASS_INIT_FIELDSET = '.qor-fieldset--init',
        CLASS_VARIANT_FEILD = '.qor-fieldset:not(.qor-fieldset--new,.qor-product-init)',
        CLASS_MEDIALIBRARY_DATA = '.qor-field__mediabox-data',
        CLASS_MEDIALIBRARY_BUTTON = '.qor-product__button-save',
        CLASS_MEDIALIBRARY_BULK_BUTTON = '.qor-product__bulk-save',
        CLASS_FILTER = '.qor-product__filter-options';

    let CLASS_VISIBLE_RESOURCE_INPUT = 'input[name*="QorResource.Variations"]:visible',
        CLASS_BULK_RESOURCE_INPUT = 'input[name*="QorResource.Variations"][type!="hidden"]';

    function replaceHtml(el, html) {
        let oldEl = typeof el === "string" ? document.getElementById(el) : el,
            newEl = oldEl.cloneNode(false);
        newEl.innerHTML = html;
        oldEl.parentNode.replaceChild(newEl, oldEl);
        return newEl;
    }

    function QorProductVariants(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorProductVariants.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorProductVariants.prototype = {
        constructor: QorProductVariants,

        init: function() {
            let $element = this.$element,
                inputName = $element.data('input-name');

            if (inputName && inputName != 'QorResource.Variations') {
                CLASS_VISIBLE_RESOURCE_INPUT = `input[name*="${inputName}"]:visible`;
                CLASS_BULK_RESOURCE_INPUT = `input[name*="${inputName}"][type!="hidden"]`;
            }

            this.bind();
            this.variants = {};
            this.PrimaryInitMetaData = {};
            this.productMetas = [];
            this.templateData = [];
            this.primaryMeta = [];
            this.existVariantsID = [];
            this.primaryMetaValue = [];
            this.replicatorTemplate = [];
            this.$tbody = $element.find(CLASS_TBODY);
            this.$replicatorBtn = $element.find(CLASS_BUTTON_ADD);
            this.$fieldBlock = $element.find('.qor-product__container>.qor-field__block');
            this.hasPrimaryMeta = $element.find(CLASS_PRIMARY_META).data('has-primary-meta');
            this.BottomSheets = $('body').data('qor.bottomsheets');
            this.initMetas();
            this.initPrimaryMeta();
        },

        bind: function() {
            $document
                .on(EVENT_REPLICATOR_ADDED, this.addVariantReplicator.bind(this))
                .on(EVENT_REPLICATORS_ADDED, this.addVariantReplicators.bind(this));

            this.$element
                .on('select2:select select2:unselect', CLASS_SELECT, this.selectVariants.bind(this))
                .on(EVENT_CLICK, '.qor-product__action--edit', this.editVariant.bind(this))
                .on(EVENT_CLICK, '.qor-product__action--delete', this.deleteVariant.bind(this))
                .on(EVENT_CLICK, '.qor-product__filter a', this.filterVariant.bind(this))
                .on(EVENT_CLICK, '.qor-product__filter-actions__edit', this.bulkEditVariants.bind(this))
                .on(EVENT_CLICK, '.qor-product__filter-actions__delete', this.bulkDeleteVariants.bind(this))
                .on(EVENT_CLICK, '.qor-product__action--add', this.addBackDeletedVariants.bind(this))
                .on(EVENT_CLICK, '.qor-product__fullscreen', this.fullscreen.bind(this))
                .on(EVENT_CLICK, 'label.mdl-checkbox input:checkbox', this.showBulkEditVariantToolbar.bind(this));
        },

        unbind: function() {
            $document
                .off(EVENT_REPLICATOR_ADDED, this.addVariantReplicator.bind(this))
                .off(EVENT_REPLICATORS_ADDED, this.addVariantReplicators.bind(this));

            this.$element
                .off('select2:select select2:unselect', CLASS_SELECT, this.selectVariants.bind(this))
                .off(EVENT_CLICK, '.qor-product__action--edit', this.editVariant.bind(this))
                .off(EVENT_CLICK, '.qor-product__action--delete', this.deleteVariant.bind(this))
                .off(EVENT_CLICK, '.qor-product__filter a', this.filterVariant.bind(this))
                .off(EVENT_CLICK, '.qor-product__filter-actions__edit', this.bulkEditVariants.bind(this))
                .off(EVENT_CLICK, '.qor-product__filter-actions__delete', this.bulkDeleteVariants.bind(this))
                .off(EVENT_CLICK, '.qor-product__action--add', this.addBackDeletedVariants.bind(this))
                .off(EVENT_CLICK, '.qor-product__fullscreen', this.fullscreen.bind(this))
                .off(EVENT_CLICK, 'label.mdl-checkbox input:checkbox', this.showBulkEditVariantToolbar.bind(this));
        },

        initMetas: function() {
            let productMetas = this.$element.find('.qor-product__meta');

            for (let i = 0, len = productMetas.length; i < len; i++) {
                this.productMetas.push($(productMetas[i]).data('inputName'));
            }

            this.setTemplate();
        },

        fullscreen: function() {
            this.$element.toggleClass('fullscreen').find('.qor-product__fullscreen i').toggle();
            $('.qor-slideout').toggleClass('overflow-hidden');
        },

        collectExistVariantsID: function() {
            let trs = this.$tbody.find(`tr:not(.${CLASS_IS_REMOVE})`);

            for (let i = 0, len = trs.length; i < len; i++) {
                this.existVariantsID.push($(trs[i]).attr('variants-id'));
            }
        },

        initPrimaryMeta: function() {
            let primaryMeta = this.$element.find(CLASS_SELECT_TYPE),
                collections = this.$element.find(CLASS_INIT_FIELDSET),
                PrimaryInitMetaData = {},
                lastObj = [],
                alreadyHaveMeta;

            for (let i = 0, len = primaryMeta.length; i < len; i++) {
                let metaArr = [],
                    meta = $(primaryMeta[i]).attr('data-filter-meta', i).data('variant-type');

                this.primaryMeta.push(meta);

                if (!collections.length) {
                    continue;
                }

                for (let j = 0, len2 = collections.length; j < len2; j++) {
                    let $collection = $(collections[j]),
                        $input = $collection.find(`[name$=${meta}]`).not('[type="hidden"]'),
                        id = $input.val(),
                        obj = {},
                        elementObj = {};

                    if ($input.is('select')) {

                        if (id) {
                            let texts = $input.find(`option[value="${id}"]`).html();

                            obj[meta] = texts;
                            obj.id = id;

                            alreadyHaveMeta = this.checkSameObj(lastObj, obj);

                            if (!alreadyHaveMeta) {
                                metaArr.push(obj);
                                this.primaryMetaValue.push({
                                    'type': texts,
                                    'meta': meta
                                });
                            }
                            // add id for old variants, will keep old collection if already have collections;
                            elementObj[`${meta}_ID`] = obj.id;
                            elementObj[meta] = obj[meta];
                            $collection.data(`variants-${meta}`, elementObj);
                        }

                    } else {
                        // handle isn't select
                    }
                    lastObj.push(obj);
                }

                metaArr.length && (PrimaryInitMetaData[`${meta}s`] = metaArr);
            }

            this.variants = this.PrimaryInitMetaData = PrimaryInitMetaData;
            this.variantsKey = this.collectObjectKeys();
            this.handleTemplateData();
            this.initPrimarySelector();
            this.primaryMetaValue.length && this.initFilter();

            this.hasPrimaryMeta && this.setCollectionID(collections);
        },

        initPrimarySelector: function() {
            let data = this.PrimaryInitMetaData,
                keys = Object.keys(data);

            this.ingoreInitChange = true;

            for (let i = 0, len = keys.length; i < len; i++) {
                let key = keys[i].slice(0, -1),
                    $select = $(`[data-variant-type="${key}"]`).find('.qor-field__input-selector'),
                    obj = data[keys[i]];

                for (let j = 0, len2 = obj.length; j < len2; j++) {

                    let option = $select.find(`option[value="${obj[j].id}"]`);

                    if (option.length) {
                        option.prop('selected', true);
                    } else {
                        $select.append(`<option selected value='${obj[j].id}'>${obj[j][key]}</option>`);
                    }

                }

                $select.trigger('change');
            }

            this.ingoreInitChange = false;

        },

        initFilter: function() {
            let primaryMetaValue = this.primaryMetaValue,
                $filter = this.$element.find(CLASS_FILTER);

            $filter.html('');

            for (let i = 0, len = primaryMetaValue.length; i < len; i++) {
                let data = primaryMetaValue[i];
                data.number = this.primaryMeta.indexOf(data.meta);

                $filter.append(window.Mustache.render(QorProductVariants.TEMPLATE_FILTER, data));
            }
        },

        toggleCheckbox: function(ele, isChecked) {
            ele.each(function() {
                if (isChecked) {
                    $(this).addClass('is-checked').find('.mdl-checkbox__input').prop('checked', true).closest('tr').addClass('is-selected');
                } else {
                    $(this).removeClass('is-checked').find('.mdl-checkbox__input').prop('checked', false).closest('tr').removeClass('is-selected');
                }
            });
        },

        filterVariant: function(e) {
            let $filter = $(e.target),
                type = $filter.data('filter-type'),
                $table = this.$element.find(CLASS_TABLE),
                unselectVariants = function() {
                    let $selectedVariants = $table.find(CLASS_TR_SELECTED);
                    if ($selectedVariants.length) {
                        this.toggleCheckbox($selectedVariants.find('label.mdl-checkbox'));
                        this.toggleCheckbox($table.find('th label.mdl-checkbox'));
                    }
                }.bind(this);

            switch (type) {
                case 'all':
                    $table.find('th label.mdl-checkbox').trigger('click');
                    break;
                case 'none':
                    unselectVariants();
                    break;
                default:
                    unselectVariants();
                    this.toggleCheckbox(this.$tbody.find(`tr[variants-id*="_${this.removeSpace(type)}_"] label.mdl-checkbox`), true);
                    break;
            }

            $table.find('label.mdl-checkbox').removeClass('is-focused');
            this.showVariantToolbar();
        },

        showBulkEditVariantToolbar: function() {
            setTimeout(this.showVariantToolbar.bind(this), 1);
        },

        showVariantToolbar: function() {
            let $selectedVariants = this.$tbody.find(CLASS_TR_SELECTED),
                $actions = this.$element.find('.qor-product__filter-actions'),
                len = $selectedVariants.length,
                $bulkSelector = this.$element.find(CLASS_TABLE).find('th label.mdl-checkbox');

            if (!len) {
                $actions.hide();
                if ($bulkSelector.hasClass('is-checked')) {
                    this.toggleCheckbox($bulkSelector);
                }
            } else {
                $actions.show().find('em').html(len);
            }

        },

        bulkEditVariants: function() {
            let $form = this.initBulkVariantsForm(),
                $tr = this.$tbody.find('tr:first'),
                $editForm = this.$element.find('.qor-variants__edit'),
                colspanLen = $tr.find('td').length,
                $emptyCol = $(`<tr class="qor-variants__edit"><td class="normal" colspan=${colspanLen}></td></tr>`),
                buttonTemp = `<button type="button" class="mdl-button mdl-js-button mdl-button--raised mdl-button--colored ${CLASS_MEDIALIBRARY_BULK_BUTTON.replace('.','')}">
                                OK
                            </button>
                            <button type="button" class="mdl-button mdl-js-button qor-product__bulk-cancel">
                                Cancel
                            </button><span class="qor-product__bulk-hint"></span>`;

            if ($editForm.length) {
                return false;
            }

            $tr.before($emptyCol);

            $form
                .appendTo($emptyCol.find('td'))
                .show()
                .append(buttonTemp)
                .trigger('enable')
                .on(EVENT_CLICK, '.qor-product-icon', this.checkBulkEdit.bind(this))
                .on(EVENT_CLICK, CLASS_MEDIALIBRARY_BULK_BUTTON, this.saveBulkEdit.bind(this))
                .on(EVENT_CLICK, '.qor-product__bulk-cancel', this.removeBulkEdit.bind(this));
        },

        checkBulkEdit: function(e) {
            let $btn = $(e.target).closest('.qor-product-icon'),
                $editor = this.$element.find('.qor-variants__edit'),
                $hint = $editor.find('.qor-product__bulk-hint'),
                $items;

            $btn.toggleClass('selected');

            $items = $editor.find('.qor-product-icon.selected').length;

            if ($items) {
                $hint.html(`${$items} items will be updated`);
            } else {
                $hint.html('');
            }
        },

        removeBulkEdit: function() {
            this.$element.find('.qor-variants__edit').remove();
        },

        saveBulkEdit: function() {
            let $parents = this.$element.find('.qor-product-icon.selected').parent(),
                $selectedVariants = this.$tbody.find(CLASS_TR_SELECTED),
                bulkData = [],
                selectedVariantsID = [];

            if (!$parents.length) {
                this.$element.find('.qor-variants__edit').remove();
                return false;
            }

            if (!$selectedVariants.length) {
                window.alert('No variants select!');
                return false;
            }

            $selectedVariants.each(function() {
                selectedVariantsID.push($(this).attr('variants-id'));
            });

            $parents.each(function() {
                let $input = $(this).find(`${CLASS_BULK_RESOURCE_INPUT},${CLASS_MEDIALIBRARY_DATA}`),
                    data = {};

                if ($input.length) {
                    if ($input.is(CLASS_MEDIALIBRARY_DATA)) {
                        data.isImage = true;
                        data.isDeleted = $(CLASS_MEDIALIBRARY_DATA).data('isDeleted');
                        data.$element = $(this);
                    }
                    data.name = $input.prop('name').match(/\.\w+$/g)[0].replace('.', '');
                    data.value = $input.is(':checkbox') ? $input.is(':checked') : $input.val();
                    bulkData.push(data);
                }
            });

            this.syncBulkEditValue(bulkData, selectedVariantsID);
        },

        syncBulkEditValue: function(datas, ids) {
            let $element = this.$element;

            for (let i = 0, len = ids.length; i < len; i++) {
                let id = `fieldset[fieldset-variants-id="${ids[i]}"]`;
                for (let j = 0, len2 = datas.length; j < len2; j++) {
                    let data = datas[j],
                        name = data.name,
                        value = data.value,
                        $td = this.$tbody.find(`${CLASS_TR_SELECTED} td[data-variant-type="${name}"]`),
                        $collectionInput = $element.find(`${id} [name$=".${name}"][type!="hidden"]`);

                    if (data.isImage) {
                        this.bulkAddImages(data, id);
                    } else {
                        $td.html(value);
                    }

                    if ($collectionInput.is(':checkbox')) {
                        $collectionInput.prop('checked', value);
                        if (value) {
                            $collectionInput.closest('label').addClass('is-checked');
                        } else {
                            $collectionInput.closest('label').removeClass('is-checked');
                        }
                    } else {
                        $collectionInput.val(value);
                    }

                }
            }

            $element.find('.qor-variants__edit').remove();

        },

        bulkAddImages: function(data, id) {
            let $element = data.$element;

            if (!$element) {
                return;
            }

            let $collection = this.$element.find(id),
                $list = $collection.find('.qor-field__mediabox-list'),
                $template = $element.find('.qor-field__mediabox-item').not('.is_deleted'),
                $td = this.$tbody.find(`${CLASS_TR_SELECTED} td[data-variant-type="${data.name}"]`),
                $img = '';

            if ($template.length) {
                $list.find('.qor-field__mediabox-item').remove();

                $template.each(function() {
                    let $this = $(this),
                        key = $this.data("primary-key"),
                        url = $this.data("original-url");

                    $img = `${$img}<img data-primary-key='${key}' src='${url}'/>`;
                });

                $td.html($img);
                $template.appendTo($list);
            } else {
                $td.html('');
                $list.find('.qor-field__mediabox-item').remove();
            }

        },

        bulkDeleteVariants: function(e) {
            let trs = this.$tbody.find(CLASS_TR_SELECTED),
                _this = this,
                confirmData = $(e.target).data();

            if (trs.length && confirmData.confirm) {
                window.QOR.qorConfirm(confirmData, function(confirm) {
                    if (confirm) {
                        for (let i = 0, len = trs.length; i < len; i++) {
                            let id = $(trs[i]).attr('variants-id');
                            _this.hideRemovedVariants(id);
                        }
                        _this.$element.find('.qor-product__filter-actions').hide();
                        _this.showVariantToolbar();
                    }
                });
            }

            return false;
        },

        initBulkVariantsForm: function() {
            let primaryMeta = this.primaryMeta,
                $form = this.$element.find('.qor-fieldset--new').clone();

            primaryMeta.push('Product');

            for (let i = 0, len = primaryMeta.length; i < len; i++) {
                $form.find(`[name$=${primaryMeta[i]}]`).not('[type="hidden"]').closest('.qor-form-section').remove();
            }

            $form.find('.qor-form-section .qor-field').prepend('<span class="qor-product-icon"><i class="material-icons normal">panorama_fish_eye</i><i class="material-icons selected">check_circle</i></span>');
            $form.find('.qor-fieldset__delete').remove();
            $form.prepend('<h2>Bulk Edit</h2>');

            return $form;
        },

        addBackDeletedVariants: function(e) {
            let $tr = $(e.target).closest('tr'),
                id = $tr.attr('variants-id');

            if (!id) {
                return false;
            }
            this.addBackVariants(id);
        },

        addBackVariants: function(id) {
            let $tr = this.$tbody.find(`tr[variants-id="${id}"]`),
                $collection = this.$element.find(`fieldset[fieldset-variants-id="${id}"]`),
                isDeleted = $tr.hasClass('is_deleted'),
                $insertRow;

            if (!$tr.length || !$collection.length) {
                return;
            }

            $tr.removeClass(`${CLASS_SHOULD_REMOVE} ${CLASS_IS_REMOVE} is_deleted is-selected`);

            $insertRow = this.$tbody.find(`tr:not(${CLASS_IS_DELETED}):last`);

            if (isDeleted) {
                $tr.find('label').removeClass('is-disabled').show().find('.mdl-checkbox__input').prop('disabled', false);
                $tr.find('.qor-product__action--add,.qor-product__action').toggle();

                $insertRow.length ? $insertRow.after($tr) : $tr.appendTo(this.$tbody);
                this.hiddenVariantsID = _.without(this.hiddenVariantsID, id);
            }

            $collection
                .removeClass(`${CLASS_SHOULD_REMOVE} ${CLASS_IS_REMOVE}`)
                .find('.qor-fieldset__alert').remove();

        },

        setCollectionID: function(collections) {
            let primaryMeta = this.primaryMeta,
                initVariantData = [];

            if (!collections.length) {
                return;
            }

            for (let i = 0, len = collections.length; i < len; i++) {
                let $collection = $(collections[i]),
                    obj = {},
                    objValues,
                    variantID;

                if ($collection.attr('fieldset-variants-id')) {
                    return;
                }

                for (let j = 0, len2 = primaryMeta.length; j < len2; j++) {
                    let variantData = $collection.data(`variants-${primaryMeta[j]}`);
                    if (variantData) {
                        obj = Object.assign({}, obj, variantData);
                    }
                }

                objValues = _.values(obj).map(this.removeSpace).sort();
                variantID = `${ID_VARIANTS_PRE}${objValues.join('_')}_`;
                obj.variantID = variantID;
                $collection.attr('fieldset-variants-id', variantID);
                initVariantData.push(obj);
            }

            this.setTableID(initVariantData);
        },

        setTableID: function(data) {
            let primaryMeta = this.primaryMeta,
                tr = this.$element.find(CLASS_TR),
                targetObj;

            for (let i = 0, len = tr.length; i < len; i++) {
                let obj = {},
                    $tr = $(tr[i]);

                for (let j = 0, len2 = primaryMeta.length; j < len2; j++) {
                    let key = primaryMeta[j],
                        innerObj = {},
                        value = $.trim($tr.find(`[data-variant-type="${key}"]`).data('variant-value'));

                    if (value) {
                        innerObj[key] = value;
                        obj = Object.assign({}, obj, innerObj);
                    } else {
                        continue;
                    }

                }

                targetObj = _.where(data, obj);
                targetObj.length && $tr.attr('variants-id', targetObj[0].variantID);
            }
        },

        removeSpace: function(value) {
            return value.toString().replace(/\s/g, '');
        },

        checkSameObj: function(obj, target) {
            let isSame;
            isSame = obj.some(function(element) {
                return _.isEqual(element, target);
            });

            return isSame;
        },

        collectObjectKeys: function(obj) {
            let keys = [],
                objs = obj || this.variants;

            keys = Object.keys(objs).filter(function(key) {
                return objs[key].length > 0;
            });

            return keys;
        },

        setTemplate: function() {
            let productMetas = this.productMetas,
                templateStart = '<tr variants-id=[[variantID]]>',
                templateEnd = `<td>
                                <button type="button" class="mdl-button mdl-js-button qor-product__action--add" style="display: none;">
                                    Add
                                </button>
                                <button type="button" id="qor-product-actions-for-[[variantID]]" class="mdl-button mdl-js-button mdl-button--icon qor-product__action">
                                    <i class="material-icons">more_vert</i>
                                </button>
                                <ul class="mdl-menu mdl-menu--bottom-right mdl-js-menu" for="qor-product-actions-for-[[variantID]]">
                                    <li class="mdl-menu__item" qor-icon-name="Edit">
                                        <a href="javascript://" class="qor-product__action--edit">Edit</a>
                                    </li>
                                    <li class="mdl-menu__item" qor-icon-name="Delete">
                                        <a href="javascript://" class="qor-product__action--delete">Delete</a>
                                    </li>
                                </ul>
                            </td></tr>`;


            _.each(productMetas, function(productMeta) {
                templateStart = `${templateStart}<td data-variant-type=${productMeta} class="mdl-data-table__cell--non-numeric">[[${productMeta}]]</td>`;
            });

            this.template = `${templateStart}${templateEnd}`;
        },

        deleteVariant: function(e) {
            let $target = $(e.target),
                id = $target.closest('tr').attr('variants-id'),
                _this = this,
                confirmData = $target.data();


            if (confirmData.confirm) {
                window.QOR.qorConfirm(confirmData, function(confirm) {
                    if (confirm) {
                        _this.hideRemovedVariants(id);
                        _this.showVariantToolbar();
                    }
                });
            }
            return false;
        },

        editVariant: function(e) {
            let $tr = $(e.target).closest('tr'),
                colspanLen = $tr.find('td').length,
                variantID = $tr.attr('variants-id'),
                $item,
                $emptyCol = $(`<tr class="qor-variant__edit"><td class="normal" colspan=${colspanLen}></td></tr>`),
                buttonTemp = `<button type="button" class="mdl-button mdl-js-button mdl-button--raised ${CLASS_MEDIALIBRARY_BUTTON.replace('.','')}">OK</button>`;

            if ($tr.next('tr.qor-variant__edit').length) {
                return false;
            }

            if (variantID) {
                $item = $(`fieldset[fieldset-variants-id="${variantID}"]`);
            } else {
                return false;
            }

            $tr.addClass(CLASS_IS_CURRENT).after($emptyCol);

            $item
                .appendTo($emptyCol.find('td'))
                .trigger('enable')
                .find(CLASS_MEDIALIBRARY_BUTTON).remove().end()
                .append(buttonTemp)
                .show().removeClass('hidden')
                .on(EVENT_KEYUP, CLASS_VISIBLE_RESOURCE_INPUT, this.syncCollectionToVariant.bind(this))
                .on(EVENT_CLICK, CLASS_MEDIALIBRARY_BUTTON, this.saveCollevtionEdit.bind(this))
                .on(EVENT_CHANGED_MEDIALIBRARY, CLASS_MEDIALIBRARY_DATA, this.syncCollectionToVariant.bind(this));

            this.hidePrimaryMeta($item);
        },

        hidePrimaryMeta: function($item) {
            let primaryMeta = this.primaryMeta;
            // hide variant primary property
            for (let i = 0, len = primaryMeta.length; i < len; i++) {
                $item.find(`[name$=${primaryMeta[i]}]`).not('[type="hidden"]').closest('.qor-form-section').hide();
            }
        },

        syncCollectionToVariant: function(e) {
            let $target = $(e.target),
                value = $target.val(),
                collectionID = $target.closest(CLASS_FIELDSET).attr('fieldset-variants-id'),
                $editableVariant = $(`tr[variants-id="${collectionID}"]`),
                variantType = $target.prop('name').match(/\.\w+$/g)[0].replace('.', ''),
                $td,
                imageValue;

            $td = $editableVariant.find(`[data-variant-type="${variantType}"]`);

            if ($target.is('textarea')) {
                imageValue = JSON.parse(value);

                if (imageValue.length) {
                    let img = '';

                    for (let i = 0, len = imageValue.length; i < len; i++) {
                        img = `${img}<img src="${imageValue[i].Url}"/>`;
                    }

                    $td.html(img);

                } else {
                    $td.html('');
                }
            } else {
                $td.html(value);
            }

        },

        saveCollevtionEdit: function(e) {
            let $target = $(e.target),
                $editableCollection = $target.closest(CLASS_FIELDSET),
                $editableVariant = $editableCollection.closest('tr');

            $editableVariant.prev().removeClass(CLASS_IS_CURRENT).end().remove();
            $editableCollection
                .appendTo(this.$fieldBlock)
                .off(EVENT_KEYUP, CLASS_VISIBLE_RESOURCE_INPUT, this.syncCollectionToVariant.bind(this))
                .off(EVENT_CLICK, CLASS_MEDIALIBRARY_BUTTON, this.saveCollevtionEdit.bind(this))
                .off(EVENT_CHANGED_MEDIALIBRARY, CLASS_MEDIALIBRARY_DATA, this.syncCollectionToVariant.bind(this))
                .hide();
        },

        selectVariants: function(e) {
            let $parent = $(e.target).closest(CLASS_SELECT_TYPE),
                type = $parent.data('variant-type'),
                params = e.params.data,
                isSelected = params.selected,
                variantValue = params.text || params.title || params.Name,
                id = params.id,
                topType = `${type}s`,
                isLastOne = !$parent.find('.select2-selection__choice').length;

            if (this.ingoreInitChange) {
                return false;
            }

            // if already have variants:
            this.variants[topType] = this.variants[topType] || [];

            if (isSelected) {
                this.doSelelct(variantValue, topType, type, id);
            } else {
                this.doUnselelct(variantValue, topType, type, id, isLastOne);
            }

            this.replicatorTemplate = [];
            this.initFilter();
        },

        doUnselelct: function(variantValue, topType, type, id, isLastOne) {
            // TODO: if no variants meta selected, should hide all.
            this.variants[topType] = this.variants[topType].filter(function(item) {
                return item[type] != variantValue;
            });

            this.removeVariants(variantValue, id, type);

            if (isLastOne) {
                this.renderVariants();
            } else {
                this.handleTemplateData();
            }

            // for bulk edit selector
            this.primaryMetaValue = _.reject(this.primaryMetaValue, function(obj) {
                return _.isMatch(obj, {
                    'type': variantValue,
                    'meta': type
                });
            });
        },

        doSelelct: function(variantValue, topType, type, id) {
            let variantData = {};
            variantData[type] = variantValue;
            variantData.id = id.toString();
            this.variants[topType].push(variantData);
            this.renderVariants();
            // for bulk edit selector
            this.primaryMetaValue.push({
                'type': variantValue,
                'meta': type
            });
        },

        removeVariants: function(value, id, type) {
            let templateDatas = this.templateData,
                data = {};

            data[type] = value;
            data[`${type}s_ID`] = id;

            for (let i = 0, len = templateDatas.length; i < len; i++) {
                let templateData = templateDatas[i];

                if (_.isMatch(templateData, data)) {
                    this.hideRemovedVariants(templateData.variantID);
                } else {
                    continue;
                }
            }
            this.showVariantToolbar();
        },

        hideRemovedVariants: function(id) {
            let $tr = this.$tbody.find(`tr[variants-id="${id}"]`),
                $editor = this.$tbody.find('.qor-variant__edit'),
                $collection = this.$element.find(`fieldset[fieldset-variants-id="${id}"]`);

            if (!$tr.length || !$collection.length || $tr.hasClass('is_deleted')) {
                return;
            }

            this.hiddenVariantsID = this.hiddenVariantsID || [];
            this.hiddenVariantsID.push(id);

            if ($editor.length) {
                this.$element.find('.qor-fieldset--new').before($editor.find('.qor-fieldset'));
                $editor.remove();
            }

            $tr
                .addClass(`${CLASS_IS_REMOVE} is_deleted`).removeClass('is-selected')
                .find('.qor-product__action--add,.qor-product__action').toggle();

            $tr
                .find('label').removeClass('is-checked').addClass('is-disabled').hide()
                .find('.mdl-checkbox__input').prop({
                    'checked': false,
                    'disabled': true
                });

            $tr.appendTo(this.$tbody);

            $collection
                .addClass(CLASS_IS_REMOVE)
                .find('.qor-fieldset__alert').remove().end()
                .find('.qor-fieldset__delete').trigger('click').hide();

        },

        renderVariants: function() {
            let variantsKey;

            variantsKey = this.collectObjectKeys();

            if (variantsKey.length === 0) {
                // empty table if no variants selected
                this.$tbody.find('tr:not(.qor-product-init)').hide();
                return;
            }

            this.variantsKey = variantsKey;
            this.convertVariantsData();

        },

        convertVariantsData: function() {
            this.handleTemplateData();
            this.renderVariantsTable();
        },

        handleTemplateData: function() {
            let maxIndices = [],
                variantsKey = this.variantsKey,
                variants = this.variants;

            this.templateData = [];

            _.each(variantsKey, function(key) {
                maxIndices.push(variants[key].length);
            });

            if (variantsKey.length === 1) {
                let variant = variants[variantsKey[0]];

                for (let i = 0, len = variant.length; i < len; i++) {
                    let item = variant[i],
                        key = _.keys(item)[0],
                        value = item[key],
                        objValues,
                        obj = {};

                    obj[key] = value;
                    obj.id = item.id;
                    objValues = _.values(obj).map(this.removeSpace).sort();
                    obj.variantID = `${ID_VARIANTS_PRE}${objValues.join('_')}_`;
                    this.templateData.push(obj);
                }

            } else {
                this.handleMultipleVariantsData(maxIndices, this.generateData.bind(this));
            }
        },

        renderVariantsTable: function() {
            let $table = this.$element.find(CLASS_TABLE),
                newObjs;

            $table
                .removeClass('is-upgraded').removeAttr('data-upgraded')
                .find('tr td:first-child,tr th:first-child').remove();

            newObjs = this.checkTemplateData().newObjs;
            $table.trigger('enable').find('.is_deleted label.mdl-checkbox').hide();

            this.$tbody.find(`.${CLASS_SHOULD_REMOVE}`).remove();

            if (newObjs.length) {
                setTimeout(function() {
                    this.doReplicator(newObjs);
                }.bind(this), 500);

            }
        },

        doReplicator: function(newObjs) {
            let $element = this.$element;

            this.replicator = this.replicator || $element.find(CLASS_FIELDSET_CONTAINER).data(NAME_REPLICATOR);
            this.replicator.add(null, newObjs, true);

            this.$element
                .find(`.${CLASS_SHOULD_REMOVE}${CLASS_VARIANT_FEILD} .qor-fieldset__delete`)
                .trigger('click').hide();
        },

        checkTemplateData: function() {
            let templateData = this.templateData,
                hiddenVariantsID = this.hiddenVariantsID || [],
                newObjs = [],
                oldObjs = [];

            this.collectExistVariantsID();
            this.$element
                .find(`${CLASS_VARIANT_FEILD},${CLASS_TR}:not(.qor-product-init)`).addClass(CLASS_SHOULD_REMOVE);

            for (let i = 0, len = templateData.length; i < len; i++) {
                let data = templateData[i],
                    hasOldData,
                    hasExistData,
                    id = data.variantID;

                if (hiddenVariantsID.length) {
                    hasOldData = _.contains(hiddenVariantsID, id);
                }

                if (this.existVariantsID.length) {
                    hasExistData = _.contains(this.existVariantsID, id);
                }

                if (hasOldData || hasExistData) {
                    oldObjs.push(data);
                    this.addBackVariants(id);
                } else {
                    this.$tbody.prepend(window.Mustache.render(this.template, data));
                    newObjs.push(data);
                }
            }

            return {
                'oldObjs': oldObjs,
                'newObjs': newObjs
            };
        },

        generateData: function(arrs) {
            let variantsKey = this.variantsKey,
                variants = this.variants,
                objValues,
                obj = {};
            // assume has Variants Data:
            // varints = {Color: [{Color: Blue},{Color: White}], Size: [{Size: S}, {Size: M}]}
            // arrs will be (2X2): [0,0],[0,1],[1,0],[1,1]
            // variantsKey = [Color, Size];
            // obj will be [{Color: Blue, Size: S},{Color: Blue, Size: M},{Color: White, Size: S},{Color: White, Size: M}]
            //
            // if have 3(or more) variants type: [Color, Size, Material]
            // varints = {Color: [{Color: Blue},{Color: White}], Size: [{Size: S}, {Size: M}], Material: [{Material: Jersey}, {Material: Cashmere}]}
            // arrs will be (2X2X2) : [0,0,0],[0,0,1],[0,1,0],[0,1,1],[1,0,0],[1,0,1],[1,1,0],[1,1,1]
            //
            // variants[variantsKey[i]][arrs[i]]);
            // variantsKey[i] will get varints.Color
            // arrs[i] will get varints.Color[0] => {Color: Blue}
            for (let i = 0, len = arrs.length; i < len; i++) {
                let item = variants[variantsKey[i]][arrs[i]];

                obj[`${variantsKey[i]}_ID`] = item.id;
                obj = Object.assign({}, obj, item);
            }

            delete obj.id;

            //obj will be:
            //Color: {"Color": "White",Colors_ID: 3,Size: "M",Sizes_ID: 2,variantID: "qor_variants_id_3_White_2_M"}

            objValues = _.values(obj).map(this.removeSpace).sort();
            obj.variantID = `${ID_VARIANTS_PRE}${objValues.join('_')}_`;
            this.templateData.push(obj);
        },

        handleMultipleVariantsData: function(maxIndices, func) {
            this.convertMultipleVariantsData(maxIndices, func, [], 0);
        },

        convertMultipleVariantsData: function(maxIndices, func, args, index) {
            if (maxIndices.length == 0) {
                func(args);
            } else {
                let rest = maxIndices.slice(1);
                for (args[index] = 0; args[index] < maxIndices[0]; ++args[index]) {
                    this.convertMultipleVariantsData(rest, func, args, index + 1);
                }
            }
        },

        // sync variants data between table and replicator
        addVariantReplicator: function(e, $item, data) {
            $item = this.syncReplicatorData($item, data);
            $item.attr('fieldset-variants-id', data.variantID).hide();
            this.replicatorTemplate.push($item.prop('outerHTML'));
        },

        addVariantReplicators: function() {
            let $div = $('<div></div>'),
                $target = this.$element.find('.qor-product__block'),
                html = this.replicatorTemplate.join('');

            $div.appendTo($target);
            replaceHtml($div[0], html);
        },

        syncReplicatorData: function($item, data) {
            let keys = Object.keys(data);

            for (let i = 0, len = keys.length; i < len; i++) {
                let $input = $item.find(`[name$=${keys[i]}]`).not('[type="hidden"]'),
                    idKey, html;

                if (!$input.length) {
                    continue;
                }

                if ($input.is('select')) {
                    let option = $input.find(`option[value="${data.id}"]`);
                    if (option.length) {
                        $input.val(data.id);
                    } else {
                        idKey = `${keys[i]}s_ID`;
                        html = `<option selected value='${data.id || data[idKey]}'>${data[keys[i]]}</option>`;
                        replaceHtml($input[0], html);
                    }
                }
            }

            return $item;

        },

        addLoading: function() {
            var $loading = $(QorProductVariants.TEMPLATE_LOADING);
            $loading.appendTo(this.$element.find(CLASS_FIELDSET_CONTAINER)).trigger('enable');
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorProductVariants.TEMPLATE_FILTER = `<li><a href="javascript://" data-filter-type="[[type]]" data-filter-meta=[[number]]>[[type]]</a></li>`;

    QorProductVariants.TEMPLATE_LOADING = (
        `<div class="qor-product__loading">
            <div><div class="mdl-spinner mdl-js-spinner is-active qor-layout__bottomsheet-spinner"></div></div>
        </div>`
    );

    QorProductVariants.plugin = function(options) {
        return this.each(function() {
            let $this = $(this);
            let data = $this.data(NAMESPACE);
            let fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                $this.data(NAMESPACE, (data = new QorProductVariants(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        let selector = '[data-toggle="qor.variations"]';

        $document
            .on(EVENT_DISABLE, function(e) {
                QorProductVariants.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorProductVariants.plugin.call($(selector, e.target));
            })
            .on('beforeEnable.qor.slideout', '.qor-slideout', function() {
                QorProductVariants.fieldsetInits = $(selector).find('.qor-fieldset--inits').html();
                $(selector).find('.qor-fieldset--inits').html('');
            })
            .on('afterEnable.qor.slideout', '.qor-slideout', function() {
                replaceHtml($(selector).find('.qor-fieldset--inits')[0], QorProductVariants.fieldsetInits);
                $document.triggerHandler(EVENT_ENABLE);
            })
            .triggerHandler(EVENT_ENABLE);

    });
    return QorProductVariants;
});
