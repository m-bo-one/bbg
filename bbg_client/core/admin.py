from django.contrib import admin, messages

from .models import BBGUser, Tank, Stat


@admin.register(Stat)
class StatAdmin(admin.ModelAdmin):

    list_display = ('tank', 'event', 'created_at')


@admin.register(BBGUser)
class BBGUserAdmin(admin.ModelAdmin):
    pass


def make_tank_resurection(modeladmin, request, queryset):
    messages.info(request, 'Tanks successfully resurected.')
    [obj.resurect() for obj in queryset]


make_tank_resurection.short_description = "Resurect selected tanks"


@admin.register(Tank)
class TankAdmin(admin.ModelAdmin):

    actions = [make_tank_resurection]

    change_form_template = 'admin/core/tank_change_form.html'

    list_display = ('name', 'tkey', 'colored_health', 'lvl', 'kda',
                    'created_at')

    readonly_fields = ('tkey', 'lvl', 'created_at', 'kda',
                       'x', 'y', 'colored_health', 'fire_rate', 'speed',
                       'direction', 'width', 'height', 'angle', 'scores_count',
                       'gun_damage', 'gun_bullets', 'gun_distance', 'nickname',
                       'kill_count', 'death_count')

    fieldsets = (
        ('Tank info', {
            'fields': ('player', 'name', 'tkey', 'lvl', 'created_at'),
        }),
        ('Tank stats', {
            'fields': ('kda', 'scores_count', 'death_count', 'kill_count'),
        }),
        ('Tank current position', {
            'classes': ('extrapretty',),
            'fields': ('nickname',
                       'x', 'y',
                       'colored_health', 'fire_rate', 'speed', 'direction',
                       'width', 'height',
                       'angle',
                       'gun_damage', 'gun_bullets', 'gun_distance')
        }),
    )

    def clear_messages(self, request):
        storage = messages.get_messages(request)
        for _ in storage:
            pass

        if len(storage._loaded_messages) == 1:
            del storage._loaded_messages[0]

    def save_model(self, request, obj, form, change):
        if '_tankresurect' in request.POST:
            obj.resurect()
            messages.info(request, 'Tank successfully resurected.')
        obj.save()
