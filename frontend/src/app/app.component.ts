import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet, RouterLink, RouterLinkActive],
  template: `
    <div class="app-container">
      <header class="glass-panel">
        <h1>MINERVA</h1>
      </header>
      
      <main>
        <router-outlet></router-outlet>
      </main>

      <nav class="glass-panel">
        <a routerLink="/search" routerLinkActive="active">Busca</a>
        <a routerLink="/library" routerLinkActive="active">Biblioteca</a>
        <a routerLink="/queue" routerLinkActive="active">Fila</a>
        <a routerLink="/settings" routerLinkActive="active">Kindle</a>
      </nav>
    </div>
  `,
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  title = 'minerva';
}
