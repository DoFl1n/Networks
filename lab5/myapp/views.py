from django.shortcuts import render
from django.contrib.auth.mixins import LoginRequiredMixin
from django.views import View

class Index(LoginRequiredMixin, View):
    def get(self, request):
        return render(request, "myapp1/index.html")